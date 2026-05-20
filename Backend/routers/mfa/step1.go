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

const step1FileName = "routers/mfa/step1"

func Step1(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessToken(ctx)

	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT EXISTS(SELECT 1 FROM devices WHERE user_id = $1 AND device_id = $2)", access.UserID, access.DeviceID).Scan(&exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Session exist check failed: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	if !exists {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Session does not exist")

		CookieProcessor.DetachAuthCookies(ctx)
		CookieProcessor.DetachMFACookies(ctx)
		CookieProcessor.DetachSSOCookies(ctx)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}

	var mail string
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT mail, blocked FROM users WHERE user_id = $1 LIMIT 1`, access.UserID).Scan(&mail, &blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account is blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	step2code, retry, err := OTPProcessor.Send(mail, MailModels.MFAInitiated, ctx.IP())
	if step2code == "" {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}

	token, err := TokenProcessor.CreateMFAToken(ctx, access.UserID, access.DeviceID, step2code, false)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}
