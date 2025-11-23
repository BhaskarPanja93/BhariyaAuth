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
	now := time.Now().UTC()
	refresh := TokenProcessor.ReadRefreshToken(ctx)
	if refresh.RefreshID == 0 {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !AccountProcessor.CheckSessionExists(refresh.UserID, refresh.RefreshID) {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Revoked session [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"This session has been revoked... Please login again"},
		})
	}
	if now.After(refresh.RefreshExpiry) {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Expired session [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"This session has expired... Please login again"},
		})
	}
	var currentIndex uint16
	err := Stores.MySQLClient.QueryRow("SELECT count FROM activities WHERE refresh = ? AND uid = ?", refresh.RefreshID, refresh.UserID).Scan(&currentIndex)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[ProcessRefresh] Unable to fetch count for [UID-%d-RID-%d] reason: %s", refresh.UserID, refresh.RefreshID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire session (DB-read issue)... Retrying"},
		})
	}
	if currentIndex != refresh.RefreshIndex {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Used old index [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	token, ok := TokenProcessor.CreateRenewToken(refresh)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
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
		Reply:      true,
	})
}

func ProcessLogout(ctx fiber.Ctx) error {
	if !TokenProcessor.MatchCSRF(ctx, TokenProcessor.ReadRefreshToken(ctx)) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	ResponseProcessor.DetachMFACookies(ctx)
	ResponseProcessor.DetachAuthCookies(ctx)
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:       true,
		Reply:         true,
		ModifyAuth:    true,
		Notifications: []string{"Logged Out Successfully"},
	})
}
