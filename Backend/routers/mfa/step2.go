package mfa

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
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

const step1FileName = "routers/mfa/step1"

// Step2 completes the Multi-Factor Authentication (MFA) process by verifying the OTP
// and upgrading the session to a "verified" state.
//
// This function finalizes MFA by validating the OTP sent in Step1 and issuing
// a verified MFA token as a cookie. It ensures that:
//  1. The client-provided MFA token is valid and untampered.
//  2. The token belongs to the MFA flow (not reused from another context).
//  3. The OTP provided matches the one issued earlier.
//  4. The associated user account exists and is not blocked.
//  5. A verified MFA token is issued and stored as a cookie for subsequent requests.
//
// Flow Summary:
//
//	validate form → decrypt token → validate token → validate OTP → verify user → mark verified → encrypt → attach cookie
//
// Security Considerations:
// - Encrypted token prevents tampering with MFA state.
// - TokenType validation prevents cross-flow token reuse.
// - OTP validation ensures user presence and control of registered email.
// - Rate limiting protects against OTP brute-force attacks.
// - Verified MFA cookie acts as a short-lived proof of second-factor authentication.
//
// Request Body:
// - token (string) → encrypted MFA payload from Step1
// - verification (string) → OTP received via email
//
// Returns:
// - 200 OK with success (MFA completed)
// - 200 OK with notification (OTP incorrect, account blocked)
// - 422 Unprocessable Entity (invalid input)
// - 500 Internal Server Error (system failures)
func Step2(ctx fiber.Ctx) error {

	// Parse incoming form data (token + OTP)
	form := new(FormModels.MFAForm)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Invalid form")
		// Penalize malformed requests
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Reconstruct MFA state from decrypted token
	data, err := TokenProcessor.ReadMFAToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())
		// Prevent token misuse across flows
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(data.UserID))+" "+strconv.Itoa(int(data.DeviceID)))

	// Validate OTP provided by user
	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Incorrect OTP")
		// High penalty to prevent brute-force attempts
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}

	// Verify user account status
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT blocked FROM users WHERE user_id = $1 LIMIT 1", data.UserID).Scan(&blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account not found")

		// Account does not exist (edge case)
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account blocked check failed: "+err.Error())

		// Unexpected database error
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Reject MFA completion if account is blocked
	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account is blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	// Upgrade MFA state: mark as verified
	data.Verified = true
	data.Step2Code = ""

	// Record verification timestamp (used for expiry/validation in downstream flows)
	data.Created = RequestProcessor.GetRequestTime(ctx)

	// Serialize updated MFA state
	token, err := StringProcessor.EncryptInterfaceToB64(data)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token create failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	// Attach MFA cookie to mark session as second-factor verified
	CookieProcessor.AttachMFACookie(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
		})
}
