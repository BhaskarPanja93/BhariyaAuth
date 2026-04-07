package mfa

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	CookieProcessor "BhariyaAuth/processors/cookies"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v3"
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
	refresh, err := TokenProcessor.ReadRefreshToken(ctx)

	// Validate refresh token presence, freshness and CSRF protection
	// CSRF validation ensures token is not replayed from another origin
	if err != nil || !TokenProcessor.VerifyCSRF(ctx, refresh) || !TokenProcessor.RefreshIsFresh(ctx, refresh) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	// Verify that the session/device still exists in database
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT EXISTS(SELECT 1 FROM devices WHERE user_id = $1 AND device_id = $2)", refresh.UserID, refresh.DeviceID).Scan(&exists)
	if err != nil {
		// Database read failure
		RateLimitProcessor.Add(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Reject if session has been revoked or removed
	if !exists {
		// Clean up all auth-related cookies to enforce logout
		CookieProcessor.DetachAuthCookies(ctx)
		CookieProcessor.DetachMFACookies(ctx)
		CookieProcessor.DetachSSOCookies(ctx)
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}

	// Fetch user email and account status
	var mail string
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT mail, blocked FROM users WHERE user_id = $1 LIMIT 1`, refresh.UserID).Scan(&mail, &blocked)
	if errors.Is(err, sql.ErrNoRows) {
		// Account does not exist (edge case)
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		// Unexpected DB failure
		RateLimitProcessor.Add(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Prevent MFA initiation for blocked users
	if blocked {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	// Send OTP to user's registered email for MFA verification
	step2code, retry := OTPProcessor.Send(mail, MailModels.MFAInitiated, ctx.IP())
	if step2code == "" {
		// OTP dispatch failed (rate-limited or provider issue)
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}

	// Construct MFA state payload for Step2
	token, err := TokenProcessor.CreateMFAToken(ctx, refresh.UserID, refresh.DeviceID, step2code)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	// Return encrypted MFA token
	// Client must submit this token in Step2 along with OTP
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}
