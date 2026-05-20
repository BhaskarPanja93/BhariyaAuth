package mfa

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const step2FileName = "routers/mfa/step2"

func Step2(ctx fiber.Ctx) error {

	form := new(FormModels.MFAForm)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Invalid form")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	data, err := TokenProcessor.ReadMFAToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(data.UserID))+" "+strconv.Itoa(int(data.DeviceID)))

	if data.Verified || data.Step2Code == "" {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "MFA token already verified or missing challenge")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Incorrect OTP")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}

	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT blocked FROM users WHERE user_id = $1 LIMIT 1", data.UserID).Scan(&blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account not found")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account blocked check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}

	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account is blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}

	data.Verified = true
	data.Step2Code = ""

	data.Created = RequestProcessor.GetRequestTime(ctx)

	token, err := StringProcessor.EncryptInterfaceToB64(data)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Token create failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	CookieProcessor.AttachMFACookie(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
		})
}
