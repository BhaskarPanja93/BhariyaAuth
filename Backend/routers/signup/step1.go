package signup

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	FormProcessor "BhariyaAuth/processors/form"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"

	"github.com/gofiber/fiber/v3"
)

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
	// - name format
	// - email format
	// - password strength
	if !FormProcessor.ReadFormData(ctx, form) ||
		!StringProcessor.NameIsValid(form.Name) ||
		!StringProcessor.EmailIsValid(form.Mail) ||
		!StringProcessor.PasswordIsStrong(form.Password) {

		// Penalize invalid attempts (moderate penalty)
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute allowed

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Check if an account already exists for the given email
	var exists bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, form.Mail).Scan(&exists)
	if err != nil {
		// Database read failure (low penalty, not user fault necessarily)
		RateLimitProcessor.Add(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Reject registration if account already exists
	if exists {
		// High penalty to discourage enumeration attacks
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountPresent},
			})
	}

	// Send OTP to user's email for verification in Step2
	step2code, retryAfter := OTPProcessor.Send(form.Mail, MailModels.SignUpStarted, ctx.IP())
	if step2code == "" {
		// OTP dispatch failed (temporary issue or abuse protection)
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retryAfter.Seconds(), // Inform client when retry is allowed
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}

	// Construct payload required for Step2
	// This includes all validated user data + OTP reference
	token, err := TokenProcessor.CreateSignUpToken(form, step2code)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	// Return encrypted token to client.
	// Client must include this token in Step2 along with OTP for verification
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}
