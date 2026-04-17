package passwordreset

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	MailNotifier "BhariyaAuth/processors/mail"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const step2FileName = "routers/passwordreset/step2"

// Step2 completes the password reset process by validating the OTP
// and updating the user's password securely.
//
// This function finalizes password recovery by verifying ownership of the email
// (via OTP) and securely updating the user's password. It ensures that:
//  1. The client-provided token is valid, untampered, and intended for password reset.
//  2. The OTP provided matches the one issued in Step1.
//  3. The associated account exists and is not blocked.
//  4. The new password meets strength requirements and is securely hashed.
//  5. All existing sessions are invalidated after password change.
//  6. The user is notified about the password change.
//
// Flow Summary:
//
//	validate input → decrypt token → validate token → validate OTP → fetch account → hash password → update DB → revoke sessions → notify user
//
// Security Considerations:
// - Encrypted token prevents tampering with reset data.
// - TokenType validation prevents cross-flow misuse.
// - OTP validation ensures email ownership.
// - Password hashing ensures secure storage.
// - All sessions are revoked to mitigate account compromise.
// - Rate limiting prevents OTP brute-force attempts.
//
// Request Body:
// - token (string) → encrypted payload from Step1
// - verification (string) → OTP received via email
// - password (string) → new password
//
// Returns:
// - 200 OK with success (on successful reset)
// - 200 OK with notification (OTP incorrect, account blocked)
// - 422 for malformed input
// - 500 for system failures
func Step2(ctx fiber.Ctx) error {

	// Parse and validate incoming form data
	form := new(FormModels.PasswordResetForm2)

	// Validate:
	// - form parsing success
	// - password strength
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Invalid form")

		// Penalize invalid input attempts
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if !StringProcessor.PasswordIsStrong(form.Password) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "New password weak")

		// Penalize invalid input attempts
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Reconstruct reset payload
	data, err := TokenProcessor.ReadPasswordResetToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(data.UserID)))

	// Validate OTP provided by user
	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Incorrect OTP")

		// High penalty to prevent brute-force OTP attacks
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.OTPIncorrect},
		})
	}

	// Fetch account status (ensure account still exists and is usable)
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT blocked FROM users WHERE user_id = $1 LIMIT 1", data.UserID, data.MailAddress).Scan(&blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

		// Account no longer exists (edge case)
		RequestProcessor.AddRateLimitWeight(ctx, 20_000) // 30 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountNotFound},
		})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account block check failed - SQL fetch: "+err.Error())

		// Unexpected DB error
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	// Prevent password reset for blocked accounts
	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account is blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountBlocked},
		})
	}

	// Hash the new password before storing
	hash, err := StringProcessor.HashPassword(form.Password)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Password hashing failed: "+err.Error())

		// Hashing failure
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	// Update password hash in database
	_, err = Stores.SQLClient.Exec(Config.CtxBG, `UPDATE users SET pw_hash = $1 WHERE user_id = $2`, hash, data.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Update password hash failed - SQL Exec: "+err.Error())

		// DB write failure
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBWriteError},
		})
	}

	// Invalidate all active sessions for this user
	// This ensures any compromised sessions are revoked
	err = AccountProcessor.DenyAllDevicesFromRenewing(data.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Revoke all devices failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{
				Notifications.PasswordChanged,
				Notifications.RevokeFailed},
		})
	}

	// Extract device metadata from request headers
	os, device, browser := StringProcessor.ParseUA(ctx.Get("User-Agent"))

	// Notify user about password change (security alert)
	err = MailNotifier.PasswordReset(data.MailAddress, MailModels.PasswordResetComplete, ctx.IP(), os, device, browser, 2)
	if err != nil {
		Logs.RootLogger.Add(Logs.Warn, step2FileName, RequestProcessor.GetRequestId(ctx), "PasswordReset mail send failed: "+err.Error())
	}

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
