package signin

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
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// Step1 initiates the signin process by validating user identity
// and preparing authentication data for Step2 (OTP or Password verification).
//
// This function acts as the entry point for user authentication. It:
//  1. Validates input (email).
//  2. Retrieves user account metadata (ID, blocked status, password availability).
//  3. Determines the authentication method (OTP or Password).
//  4. Initiates OTP flow if required.
//  5. Constructs a secure, encrypted payload for Step2 verification.
//
// Flow Summary:
//
//	validate input → fetch user → validate account → select process → (OTP send | password check) → encrypt state → return token
//
// Supported Authentication Modes:
// - OTP: Email-based one-time password verification.
// - Password: Traditional password authentication.
//
// Security Considerations:
// - Rate limiting prevents enumeration and brute-force attempts.
// - Account existence is explicitly returned (may expose enumeration risk).
// - OTP is required for passwordless signin.
// - Stateless design: authentication state is encrypted and returned to client.
// - TokenType ensures tokens are not reused across flows.
//
// Request Body:
// - mail (string)
// - remember ("yes" | "no")
//
// Route Parameters:
// - process: "otp" | "password"
//
// Returns:
// - 200 OK with encrypted token (on success)
// - 200 OK with notifications (account issues, OTP failure, password missing)
// - 422 Unprocessable Entity (invalid input/process)
// - 500 Internal Server Error (system failures)
func Step1(ctx fiber.Ctx) error {

	// Parse and validate incoming form data
	form := new(FormModels.SignInForm1)

	// Validate:
	// - form parsing success
	// - email format correctness
	if !FormProcessor.ReadFormData(ctx, form) ||
		!StringProcessor.EmailIsValid(form.Mail) {

		// Penalize malformed input
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Fetch user metadata in a single query (optimized DB access)
	var userID int32
	var blocked bool
	var hasPassword bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id, blocked, pw_hash IS NOT NULL AND pw_hash <> '' FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID, &blocked, &hasPassword)
	if errors.Is(err, pgx.ErrNoRows) {
		// Account does not exist (high penalty to prevent enumeration abuse)
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		// Unexpected DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Prevent signin if account is blocked
	if blocked {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	// Determine authentication method from route parameter
	process := ctx.Params(ProcessParam)

	// OTP reference (used only for OTP flow)
	step2code := ""
	if process == OTPProcess {

		// Initiate OTP-based signin
		step2code, retry := OTPProcessor.Send(form.Mail, MailModels.SignInStarted, ctx.IP())
		if step2code == "" {
			// OTP dispatch failure (rate-limited or provider issue)
			RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Reply:         retry.Seconds(),
					Notifications: []string{Notifications.OTPSendFailed},
				})
		}
	} else if process == PasswordProcess {

		// Ensure password-based signin is allowed
		if !hasPassword {
			// Prevent signin if password was never set (e.g., SSO-only account)
			RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.PasswordNotSet},
				})
		}
	} else {
		// Invalid authentication mode
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Construct payload for Step2 authentication
	token, err := TokenProcessor.CreateSignInToken(form, userID, process, step2code)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	// Return encrypted token to client.
	// Client must use this token in Step2 along with password or OTP
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}
