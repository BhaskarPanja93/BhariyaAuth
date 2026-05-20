package passwordreset

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const step1FileName = "routers/passwordreset/step1"

func Step1(ctx fiber.Ctx) error {

	form := new(FormModels.PasswordResetForm1)

	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Invalid form")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if !StringProcessor.EmailIsValid(form.Mail) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Invalid email")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+form.Mail)

	var userID int32
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account not found")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountNotFound},
		})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account find failed - SQL query: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	step2code, retry, err := OTPProcessor.Send(form.Mail, MailModels.PasswordResetStarted, ctx.IP())

	if step2code == "" {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Reply:         retry.Seconds(),
			Notifications: []string{Notifications.OTPSendFailed},
		})
	}

	token, err := TokenProcessor.CreatePasswordResetToken(form.Mail, userID, step2code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}
