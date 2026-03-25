package account

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"database/sql"
	"errors"

	"time"

	"github.com/gofiber/fiber/v3"
)

func Logout(ctx fiber.Ctx) error {
	refresh, ok := TokenProcessor.ReadRefreshToken(ctx)
	if !ok || !TokenProcessor.MatchCSRF(ctx, refresh) {
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	CookieProcessor.DetachAuthCookies(ctx)
	CookieProcessor.DetachMFACookies(ctx)
	CookieProcessor.DetachSSOCookies(ctx)
	AccountProcessor.DeleteSession(refresh.UserID, refresh.DeviceID)
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
		})
}

func Refresh(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	refresh, ok := TokenProcessor.ReadRefreshToken(ctx)
	// Refresh requires CSRF
	if !ok || !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	// Expiry should be in the future for a refresh token to be called active
	if now.After(refresh.Expiry) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionExpired},
			})
	}
	// Initiate a transaction because 2 separate (read and write) actions will be performed internally
	// and concurrency may cause overwriting
	tx, err := Stores.SQLClient.Begin(Config.CtxBG)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	defer tx.Rollback(Config.CtxBG)
	// Fetch the visits column which in other words mean the most recent version of refresh token for that specific device id
	var visits int16
	err = tx.QueryRow(Config.CtxBG, "SELECT visits FROM devices where user_id = $1 AND device_id = $2 LIMIT 1 FOR UPDATE", refresh.UserID, refresh.DeviceID).Scan(&visits)
	if errors.Is(err, sql.ErrNoRows) { // Row(device) not found
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	} else if err != nil { // Any other DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// DB visits should match refresh visits
	if visits != refresh.Visits {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}
	// Create their access and refresh tokens
	token, ok := TokenProcessor.CreateFreshToken(ctx, refresh.UserID, refresh.DeviceID, refresh.UserType, refresh.RememberMe, "email-login")
	if !ok {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.UnknownError},
		})
	}
	// Increment the visit count in database, which will in turn auto-revoke all old versions of the refresh token
	_, err = tx.Exec(Config.CtxBG, "UPDATE devices SET visits = $1, updated = $2 WHERE device_id = $3 AND user_id = $4", refresh.Visits+1, now, refresh.DeviceID, refresh.UserID)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}
	// Release the row for future access
	err = tx.Commit(Config.CtxBG)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}
	// Attach the new refresh token and CSRF as cookie
	CookieProcessor.AttachAuthCookies(ctx, token)
	// Return the new access token and its expiry
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires,
		})
}
