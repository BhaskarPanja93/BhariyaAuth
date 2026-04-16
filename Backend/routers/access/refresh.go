package access

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	CookieProcessor "BhariyaAuth/processors/cookies"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const refreshFileName = "routers/access/refresh"

// Refresh issues a new access token and rotates the refresh token
// for an existing authenticated session.
//
// This function validates the current refresh token and safely rotates it
// to prevent replay attacks. It ensures that:
//  1. The refresh token is valid and CSRF-protected.
//  2. The token has not expired.
//  3. The session/device exists in the database.
//  4. The refresh token version ("visits") matches the latest stored version.
//  5. A new token pair (access + refresh) is issued.
//  6. The previous refresh token is invalidated via version increment.
//
// Flow Summary:
//
//	validate token → verify CSRF → check expiry → lock session → validate version → issue new token → rotate version → commit → return tokens
//
// Security Model:
// - Uses **rotating refresh tokens** (versioned via `visits`).
// - Prevents replay attacks: old tokens become invalid after rotation.
// - Uses DB row locking (`FOR UPDATE`) to ensure concurrency safety.
// - CSRF validation ensures refresh requests are same-origin.
// - Transaction guarantees atomic read + write.
//
// Key Concepts:
// - visits: acts as a version counter for refresh tokens per device.
// - device_id: identifies a unique session.
// - refresh token is valid only if visits matches DB value.
//
// Returns:
// - 200 OK with new access token (on success)
// - 401 Unauthorized (invalid/expired/revoked session)
// - 500 Internal Server Error (DB or system failure)
func Refresh(ctx fiber.Ctx) error {

	// Extract refresh token from request
	refresh, err := TokenProcessor.ReadRefreshToken(ctx)

	// Validate presence and CSRF protection
	if err != nil || !TokenProcessor.VerifyCSRF(ctx, refresh) || !TokenProcessor.RefreshIsFresh(ctx, refresh) {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Refresh invalid/expired/CSRF incorrect")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	Logs.RootLogger.Add(Logs.Intent, refreshFileName, RequestProcessor.GetRequestId(ctx), "Request for: "+strconv.Itoa(int(refresh.UserID))+" "+strconv.Itoa(int(refresh.DeviceID)))

	// Begin transaction to ensure atomic read-modify-write
	tx, err := Stores.SQLClient.Begin(Config.CtxBG)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Transaction create failed - SQL Begin: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	defer tx.Rollback(Config.CtxBG)

	// Fetch current session version (visits) with row-level lock
	var visits int16
	err = tx.QueryRow(Config.CtxBG, "SELECT visits FROM devices where user_id = $1 AND device_id = $2 LIMIT 1 FOR UPDATE", refresh.UserID, refresh.DeviceID).Scan(&visits)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Account not found")
		// Session does not exist → revoked or invalid
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed")
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	// Ensure token version matches latest DB version
	// Prevents replay of old refresh tokens
	if visits != refresh.Visits {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Incorrect visit count: received "+strconv.Itoa(int(refresh.Visits))+" expected "+strconv.Itoa(int(visits)))
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}

	// Increment session version to invalidate previous refresh token
	_, err = tx.Exec(Config.CtxBG, "UPDATE devices SET visits = $1, updated = $2 WHERE device_id = $3 AND user_id = $4", refresh.Visits+1, RequestProcessor.GetRequestTime(ctx), refresh.DeviceID, refresh.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Visit increment failed - SQL Exec: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	// Commit transaction (finalize rotation)
	err = tx.Commit(Config.CtxBG)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Commit failed - SQL Commit: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	// Generate new access + refresh tokens
	token, err := TokenProcessor.CreateRenewToken(ctx, refresh)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Access create failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.UnknownError},
			})
	}

	// Attach new refresh + CSRF cookies
	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return new access token and expiry
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires,
		})
}
