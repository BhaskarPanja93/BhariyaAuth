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

func Refresh(ctx fiber.Ctx) error {

	refresh, err := TokenProcessor.ReadRefreshToken(ctx)

	if err != nil || !TokenProcessor.VerifyCSRF(ctx, refresh) || !TokenProcessor.RefreshIsFresh(ctx, refresh) {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Refresh invalid/expired/CSRF incorrect")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	Logs.RootLogger.Add(Logs.Intent, refreshFileName, RequestProcessor.GetRequestId(ctx), "Request for: "+strconv.Itoa(int(refresh.UserID))+" "+strconv.Itoa(int(refresh.DeviceID)))

	tx, err := Stores.SQLClient.Begin(Config.CtxBG)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Transaction create failed - SQL Begin: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	defer tx.Rollback(Config.CtxBG)

	var visits int16
	err = tx.QueryRow(Config.CtxBG, "SELECT visits FROM devices where user_id = $1 AND device_id = $2 LIMIT 1 FOR UPDATE", refresh.UserID, refresh.DeviceID).Scan(&visits)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Account not found")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	if visits != refresh.Visits {
		Logs.RootLogger.Add(Logs.Blocked, refreshFileName, RequestProcessor.GetRequestId(ctx), "Incorrect visit count: received "+strconv.Itoa(int(refresh.Visits))+" expected "+strconv.Itoa(int(visits)))
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}

	_, err = tx.Exec(Config.CtxBG, "UPDATE devices SET visits = $1, updated = $2 WHERE device_id = $3 AND user_id = $4", refresh.Visits+1, RequestProcessor.GetRequestTime(ctx), refresh.DeviceID, refresh.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Visit increment failed - SQL Exec: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	err = tx.Commit(Config.CtxBG)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Commit failed - SQL Commit: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	token, err := TokenProcessor.CreateRenewToken(ctx, refresh)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, refreshFileName, RequestProcessor.GetRequestId(ctx), "Access create failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.UnknownError},
			})
	}

	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Info, refreshFileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires,
		})
}
