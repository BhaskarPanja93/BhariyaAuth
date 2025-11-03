package account

import (
	Config "BhariyaAuth/constants/config"
	AuthStateModels "BhariyaAuth/models/authstate"
	ResponseModels "BhariyaAuth/models/responses"
	ResponseProcessor "BhariyaAuth/processors/response"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"

	"time"

	"github.com/gofiber/fiber/v3"
)

func ProcessRefresh(ctx fiber.Ctx) error {
	refresh := TokenProcessor.ReadRefreshToken(ctx)
	if !refresh.RememberMe || TokenProcessor.ReadAccessToken(ctx).RefreshID != refresh.RefreshID || TokenProcessor.RefreshIsBlacklisted(refresh.UserID, refresh.RefreshID, true) {
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.RefreshBlocked,
			ResponseModels.DefaultAuth,
			[]string{"This token can not be refreshed, please login again"},
			nil,
			nil,
		))
	} else {
		if !TokenProcessor.MatchCSRF(ctx, refresh) || time.Since(refresh.RefreshUpdated) > Config.RefreshTokenExpireDelta {
			return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.RefreshFailed,
				ResponseModels.DefaultAuth,
				[]string{"CSRF didnt match or refresh token expired"},
				nil,
				nil,
			))
		} else {
			var currentIndex uint16
			err := Stores.MySQLClient.QueryRow("SELECT count FROM activities WHERE refresh = ? AND uid = ?", refresh.RefreshID, refresh.UserID).Scan(&currentIndex)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.RefreshFailed,
					ResponseModels.DefaultAuth,
					[]string{"Refresh Failed, please try again"},
					nil,
					nil,
				))
			}
			if currentIndex != refresh.RefreshIndex {
				return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.RefreshBlocked,
					ResponseModels.DefaultAuth,
					[]string{"Refresh token too old"},
					nil,
					nil,
				))
			} else {
				token := TokenProcessor.CreateRenewToken(refresh)
				_, err = Stores.MySQLClient.Exec("UPDATE activities SET count = ?, updated = ? WHERE refresh = ? AND uid = ?", refresh.RefreshIndex+1, time.Now(), refresh.RefreshID, refresh.UserID)
				if err != nil {
					return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.RefreshFailed,
						ResponseModels.DefaultAuth,
						[]string{"Server error, please try again"},
						nil,
						nil,
					))
				}
				ResponseProcessor.AttachAuthCookies(ctx, ResponseProcessor.GenerateAuthCookies(token))
				return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
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
		}
	}
}

func ProcessLogout(ctx fiber.Ctx) error {
	ResponseProcessor.DetachAuthCookies(ctx)
	return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
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

func Me(ctx fiber.Ctx) error {
	access := TokenProcessor.ReadAccessToken(ctx)
	refresh := TokenProcessor.ReadRefreshToken(ctx)
	timePassed := time.Since(access.AccessCreated)
	var state AuthStateModels.StateT
	if timePassed < Config.AccessTokenFreshnessExpireDelta {
		state = AuthStateModels.Fresh
	} else if timePassed > Config.AccessTokenExpireDelta {
		state = AuthStateModels.Stale
	} else if timePassed > Config.AccessTokenFreshnessExpireDelta {
		state = AuthStateModels.Active
	} else {
		state = AuthStateModels.Unknown
	}
	return ctx.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"access":  access,
		"refresh": refresh,
		"state":   state,
	})
}
