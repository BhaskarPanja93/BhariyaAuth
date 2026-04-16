package mfa

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	CookieProcessor "BhariyaAuth/processors/cookies"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// Step1 initiates a Multi-Factor Authentication (MFA) challenge for an already authenticated session.
//
// This function is used for step-up authentication (e.g., sensitive actions, session verification).
// It validates the current session using the refresh token and CSRF protection, then:
//  1. Ensures the session is valid, active, and not expired.
//  2. Confirms the device/session exists in the database.
//  3. Verifies the associated user account is valid and not blocked.
//  4. Sends an OTP to the user's registered email.
//  5. Returns a secure, encrypted MFA token for Step2 verification.
//
// Flow Summary:
//
//	validate refresh → verify CSRF → check expiry → verify session → fetch user → send OTP → encrypt state → return token
//
// Security Considerations:
// - Refresh token must pass CSRF validation (prevents token misuse).
// - Session must exist in DB (prevents usage of revoked tokens).
// - OTP ensures user presence/ownership before sensitive actions.
// - Stateless MFA token is encrypted to prevent tampering.
// - Rate limiting prevents abuse and OTP flooding.
//
// Dependencies:
// - Requires a valid refresh token cookie.
// - Requires CSRF token validation.
//
// Returns:
// - 200 OK with encrypted MFA token (on success)
// - 200 OK with notification (blocked account, OTP failure)
// - 401 Unauthorized (invalid/expired session)
// - 500 Internal Server Error (system failures)
func Step1(ctx fiber.Ctx) error {

	// Extract refresh token from request (typically cookie-based)
	access, err := TokenProcessor.ReadAccessToken(ctx)

	// Validate refresh token presence, freshness and CSRF protection
	// CSRF validation ensures token is not replayed from another origin
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))
	// Verify that the session/device still exists in database
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT EXISTS(SELECT 1 FROM devices WHERE user_id = $1 AND device_id = $2)", access.UserID, access.DeviceID).Scan(&exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Session exist check failed: "+err.Error())
		// Database read failure
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Reject if session has been revoked or removed
	if !exists {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Session does not exist")

		// Clean up all auth-related cookies to enforce logout
		CookieProcessor.DetachAuthCookies(ctx)
		CookieProcessor.DetachMFACookies(ctx)
		CookieProcessor.DetachSSOCookies(ctx)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}

	// Fetch user email and account status
	var mail string
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT mail, blocked FROM users WHERE user_id = $1 LIMIT 1`, access.UserID).Scan(&mail, &blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

		// Account does not exist (edge case)
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed: "+err.Error())

		// Unexpected DB failure
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Prevent MFA initiation for blocked users
	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account is blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	// Send OTP to user's registered email for MFA verification
	step2code, retry, err := OTPProcessor.Send(mail, MailModels.MFAInitiated, ctx.IP())
	if step2code == "" {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

		// OTP dispatch failed (rate-limited or provider issue)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}

	// Construct MFA state payload for Step2
	token, err := TokenProcessor.CreateMFAToken(ctx, access.UserID, access.DeviceID, step2code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return encrypted MFA token
	// Client must submit this token in Step2 along with OTP
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}
