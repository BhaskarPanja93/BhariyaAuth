package signup

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const step2FileName = "routers/signup/step2"

// Step2 completes the user registration process using OTP verification.
//
// This function finalizes registration by validating the OTP and securely reconstructing
// the user data generated in Step1. It ensures that:
//  1. The client-provided token is valid, untampered, and intended for registration.
//  2. The email has not been registered since Step1 (race condition protection).
//  3. The OTP provided matches the one issued earlier.
//  4. A new user account is created upon successful validation.
//  5. A device session is registered and authentication tokens are issued.
//  6. MFA state is initialized since OTP verification was already performed.
//
// Flow Summary:
//
//	validate form → decrypt token → validate token → re-check email → validate OTP → create user → create session → issue tokens → attach MFA
//
// Security Considerations:
// - Encrypted token prevents client-side tampering of registration data.
// - TokenType validation ensures token is not reused across flows.
// - Email re-check prevents race-condition account duplication.
// - OTP validation ensures ownership of email.
// - Rate limiting is applied to prevent brute-force OTP attempts.
// - MFA token is issued immediately since OTP acts as first-factor verification.
//
// Request Body:
// - token (string) → encrypted payload from Step1
// - verification (string) → OTP received via email
//
// Returns:
// - 200 OK with auth tokens on success
// - 200 OK with notifications for logical failures (OTP incorrect, account exists)
// - 422 for malformed requests
// - 500 for system failures
func Step2(ctx fiber.Ctx) error {

	form := new(FormModels.SignUpForm2)

	// Parse incoming form data (token + OTP)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		// Penalize malformed input
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Reconstruct original Step1 payload
	data, err := TokenProcessor.ReadSignUpToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested account: "+data.MailAddress)

	// Re-check if account already exists (race condition protection)
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, data.MailAddress).Scan(&exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account existence check failed - SQL query: "+err.Error())

		// Database read failure
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	// Abort if account was created after Step1
	if exists {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account exists during step2")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountPresent},
		})
	}

	// Validate OTP provided by user
	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Incorrect OTP")

		// High penalty to prevent OTP brute-force attempts
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.OTPIncorrect},
		})
	}

	// Default user type assignment
	userType := UserTypes.All.Viewer.Short

	// Create new user account in database
	userID, err := AccountProcessor.RecordNewUser(ctx, userType, data.Password, data.MailAddress, data.Name)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignUp failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBWriteError},
		})
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed Up: "+data.MailAddress+" "+strconv.Itoa(int(userID)))

	// Register user device/session
	deviceID, err := AccountProcessor.RecordReturningUser(ctx, data.MailAddress, userID, data.Remember, false)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignIn failed: "+err.Error())

		// Account created but signin partially failed
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed In: "+strconv.Itoa(int(userID)))

	// Generate authentication tokens (access + refresh)
	token, err := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, data.Remember, "email-signup")
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		// Account exists but token issuance failed
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}

	// Attach refresh token in cookies
	CookieProcessor.AttachAuthCookies(ctx, token)

	// Initialize MFA token (OTP already verified → mark as verified)
	mfaToken, err := TokenProcessor.CreateMFAToken(ctx, userID, deviceID, data.Step2Code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Warn, step2FileName, RequestProcessor.GetRequestId(ctx), "MFA creation failed: "+err.Error())
	}
	// Attach MFA cookie
	CookieProcessor.AttachMFACookie(ctx, mfaToken)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Completed request: "+strconv.Itoa(int(userID))+" "+strconv.Itoa(int(deviceID)))
	// Return access token and expiry to client
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
		NewToken:   token.AccessToken,
		Reply:      token.AccessExpires,
	})
}
