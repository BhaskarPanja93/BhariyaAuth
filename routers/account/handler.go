package account

import (
	Config "BhariyaAuth/constants/config"
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
	refresh := TokenProcessor.ReadRefreshToken(ctx)
	if refresh.RefreshID == 0 {
		RateLimitProcessor.SetValue(ctx)
		ResponseProcessor.DetachAuthCookies(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"Invalid refresh token"},
				nil,
				nil,
			))
	}
	if !TokenProcessor.MatchCSRF(ctx, refresh) {
		Logger.IntentionalFailure(fmt.Sprintf("Refresh-1 Attempted refresh without CSRF [%d-%d]", refresh.UserID, refresh.RefreshID))
		RateLimitProcessor.SetValue(ctx)
		ResponseProcessor.DetachAuthCookies(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"CSRF didnt match"},
				nil,
				nil,
			))
	}
	if !AccountProcessor.SessionExists(refresh.UserID, refresh.RefreshID) {
		Logger.IntentionalFailure(fmt.Sprintf("Refresh-1 Attempted refresh on revoked session [%d-%d]", refresh.UserID, refresh.RefreshID))
		RateLimitProcessor.SetValue(ctx)
		ResponseProcessor.DetachAuthCookies(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshBlocked,
				ResponseModels.DefaultAuth,
				[]string{"This session has been revoked, please login again"},
				nil,
				nil,
			))
	}
	if time.Since(refresh.RefreshUpdated) > Config.RefreshTokenExpireDelta {
		Logger.IntentionalFailure(fmt.Sprintf("Refresh-1 Attempted refresh on expired session [%d-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"Refresh token has expired"},
				nil,
				nil,
			))
	}
	var currentIndex uint16
	err := Stores.MySQLClient.QueryRow("SELECT count FROM activities WHERE refresh = ? AND uid = ?", refresh.RefreshID, refresh.UserID).Scan(&currentIndex)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Refresh-1 Couldnt find count from DB [%d-%d]: %s", refresh.UserID, refresh.RefreshID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"Refresh Failed, please try again"},
				nil,
				nil,
			))
	}
	if currentIndex != refresh.RefreshIndex {
		Logger.IntentionalFailure(fmt.Sprintf("Refresh-1 Attempted refresh with old index [%d-%d]", refresh.UserID, refresh.RefreshID))
		RateLimitProcessor.SetValue(ctx)
		ResponseProcessor.DetachAuthCookies(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshBlocked,
				ResponseModels.DefaultAuth,
				[]string{"Refresh token is old"},
				nil,
				nil,
			))
	}
	token := TokenProcessor.CreateRenewToken(refresh)
	_, err = Stores.MySQLClient.Exec("UPDATE activities SET count = ?, updated = ? WHERE refresh = ? AND uid = ?", refresh.RefreshIndex+1, time.Now(), refresh.RefreshID, refresh.UserID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Refresh-1 Failed to update count on DB [%d-%d] %s", refresh.UserID, refresh.RefreshID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"Server error, please try again"},
				nil,
				nil,
			))
	}
	ResponseProcessor.AttachAuthCookies(ctx, token)
	Logger.Success(fmt.Sprintf("Refresh-1 Succeeded [%d-%d]", refresh.UserID, refresh.RefreshID))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.RefreshSucceeded,
			ResponseModels.AuthT{
				Allowed: true,
				Change:  true,
				Token:   token.AccessToken,
			},
			[]string{"Refreshed Successfully"},
			nil,
			nil,
		))
}

func ProcessLogout(ctx fiber.Ctx) error {
	ResponseProcessor.DetachAuthCookies(ctx)
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.SignedOut,
			ResponseModels.AuthT{
				Allowed: true,
				Change:  true,
				Token:   "",
			},
			[]string{"Logged Out Successfully"},
			nil,
			nil,
		))
}
