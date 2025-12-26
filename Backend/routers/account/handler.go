package account

import (
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	ResponseProcessor "BhariyaAuth/processors/response"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"fmt"

	"time"

	"github.com/gofiber/fiber/v3"
)

func ProcessRefresh(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	refresh, ok := TokenProcessor.ReadRefreshToken(ctx)
	if !ok || !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if now.After(refresh.RefreshExpiry) || !AccountProcessor.CheckSessionExists(refresh.UserID, refresh.RefreshID) {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Expired/Revoked session [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		RateLimitProcessor.Set(ctx)
		ResponseProcessor.DetachAuthCookies(ctx)
		ResponseProcessor.DetachMFACookies(ctx)
		ResponseProcessor.DetachSSOCookies(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(ResponseModels.APIResponseT{
			Notifications: []string{"This session has expired..."},
		})
	}
	var currentIndex uint16
	err := Stores.MySQLClient.QueryRow("SELECT count FROM activities WHERE refresh = ? AND uid = ?", refresh.RefreshID, refresh.UserID).Scan(&currentIndex)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[ProcessRefresh] Unable to fetch count for [UID-%d-RID-%d] reason: %s", refresh.UserID, refresh.RefreshID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{"Failed to acquire session (DB-read issue)... Retrying"},
		})
	}
	if currentIndex != refresh.RefreshIndex {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Used old index [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	token, ok := TokenProcessor.CreateRenewToken(refresh, ctx)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{"Failed to acquire session (Encryptor issue)... Retrying"},
		})
	}
	_, err = Stores.MySQLClient.Exec("UPDATE activities SET count = ?, updated = ? WHERE refresh = ? AND uid = ?", refresh.RefreshIndex+1, now, refresh.RefreshID, refresh.UserID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[ProcessRefresh] Failed to update count for [UID-%d-RID-%d] reason: %s", refresh.UserID, refresh.RefreshID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire session (DB-write issue)... Retrying"},
		})
	}
	ResponseProcessor.AttachAuthCookies(ctx, token)
	Logger.Success(fmt.Sprintf("[ProcessRefresh] Success for [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
		NewToken:   token.AccessToken,
	})
}

func ProcessLogout(ctx fiber.Ctx) error {
	refresh, ok := TokenProcessor.ReadRefreshToken(ctx)
	if !ok || !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	ResponseProcessor.DetachAuthCookies(ctx)
	ResponseProcessor.DetachMFACookies(ctx)
	ResponseProcessor.DetachSSOCookies(ctx)
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
	})
}
