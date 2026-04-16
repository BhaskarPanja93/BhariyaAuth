package signup

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const step1FileName = "routers/signup/step1"

// Step1 initiates the first phase of user registration with email verification.
//
// This function validates user-provided registration data and prepares a secure payload
// for the second step of registration (OTP verification). It ensures that:
//  1. Input data is structurally and semantically valid.
//  2. The email is not already registered in the system.
//  3. An OTP is generated and sent to the provided email address.
//  4. A secure, encrypted token containing all required registration data is returned.
//
// This token must be submitted in Step2 along with the OTP to complete registration.
//
// Flow Summary:
//
//	validate input → check email uniqueness → send OTP → package state → encrypt → return token
//
// Security Considerations:
// - Input validation prevents malformed or weak credentials.
// - Rate limiting is applied aggressively to mitigate abuse/brute-force attempts.
// - OTP ensures email ownership verification.
// - Sensitive data (password, OTP reference) is encrypted before returning to client.
// - No server-side session is stored; state is fully client-carried but protected via encryption.
//
// Request Body:
// - name (string)
// - mail (string)
// - password (string)
// - remember ("yes" | "no")
//
// Returns:
// - 200 OK with encrypted token (on success)
// - 200 OK with notification (logical failure like account exists or OTP failure)
// - 422 Unprocessable Entity (invalid input)
// - 500 Internal Server Error (unexpected system failure)
func Step1(ctx fiber.Ctx) error {

	// Parse and validate incoming form data
	form := new(FormModels.SignUpForm1)

	// Validate:
	// - form parsing success
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		// Penalize invalid attempts (moderate penalty)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute allowed

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Validate:
	// - name format
	// - email format
	// - password strength
	validName := StringProcessor.NameIsValid(form.Name)
	validEmail := StringProcessor.EmailIsValid(form.Mail)
	validPassword := StringProcessor.PasswordIsStrong(form.Password)
	if !validName || !validEmail || !validPassword {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Form values invalid: "+strconv.FormatBool(validName)+" "+form.Name+" "+strconv.FormatBool(validEmail)+" "+form.Mail+" "+strconv.FormatBool(validPassword))

		// Penalize invalid attempts (moderate penalty)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute allowed

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested account: "+form.Mail)

	// Check if an account already exists for the given email
	var exists bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, form.Mail).Scan(&exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account exists: "+form.Mail)

		// Database read failure (low penalty, not user fault necessarily)
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	// Reject registration if account already exists
	if exists {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account exists: "+form.Mail)

		// High penalty to discourage enumeration attacks
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountPresent},
		})
	}

	// Send OTP to user's email for verification in Step2
	step2code, retryAfter, err := OTPProcessor.Send(form.Mail, MailModels.SignUpStarted, ctx.IP())
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

		// OTP dispatch failed
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Reply:         retryAfter.Seconds(), // Inform client when retry is allowed
			Notifications: []string{Notifications.OTPSendFailed},
		})
	}

	// Construct payload required for Step2
	// This includes all validated user data + OTP reference
	token, err := TokenProcessor.CreateSignUpToken(form, step2code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Competed request")

	// Return encrypted token to client.
	// Client must include this token in Step2 along with OTP for verification
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}
