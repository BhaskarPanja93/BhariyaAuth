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
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"

	"github.com/gofiber/fiber/v3"
)

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

	// Parse incoming form data (token + OTP)
	form := new(FormModels.SignUpForm2)
	if !FormProcessor.ReadFormData(ctx, form) {
		// Penalize malformed input
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Reconstruct original Step1 payload
	data, err := TokenProcessor.ReadSignUpToken(form.Token)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	// Re-check if account already exists (race condition protection)
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, data.MailAddress).Scan(&exists)
	if err != nil {
		// Database read failure
		RateLimitProcessor.Add(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Abort if account was created after Step1
	if exists {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountPresent},
			})
	}

	// Validate OTP provided by user
	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		// High penalty to prevent OTP brute-force attempts
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}

	// Default user type assignment
	userType := UserTypes.All.Viewer.Short

	// Create new user account in database
	userID, err := AccountProcessor.RecordNewUser(ctx, userType, data.Password, data.MailAddress, data.Name)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	// Register user device/session
	deviceID, err := AccountProcessor.RecordReturningUser(ctx, data.MailAddress, userID, data.Remember, false)
	if err != nil {
		// Account created but signin partially failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}

	// Generate authentication tokens (access + refresh)
	token, err := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, data.Remember, "email-signup")
	if err != nil {
		// Account exists but token issuance failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}

	// Attach refresh token in cookies
	CookieProcessor.AttachAuthCookies(ctx, token)

	// Initialize MFA token (OTP already verified → mark as verified)
	mfaToken, err := TokenProcessor.CreateMFAToken(ctx, userID, deviceID, data.Step2Code)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountCreated},
			})
	}

	// Attach MFA cookie
	CookieProcessor.AttachMFACookie(ctx, mfaToken)

	// Return access token and expiry to client
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires,
		})
}
