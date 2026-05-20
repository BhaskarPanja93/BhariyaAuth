package signin

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
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const step1FileName = "routers/signin/step1"

func Step1(ctx fiber.Ctx) error {

	form := new(FormModels.SignInForm1)

	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	validEmail := StringProcessor.EmailIsValid(form.Mail)
	if !validEmail {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Form values invalid: "+form.Mail)

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+form.Mail+" "+form.Process)

	var userID int32
	var blocked bool
	var hasPassword bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id, blocked, pw_hash IS NOT NULL AND pw_hash <> '' FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID, &blocked, &hasPassword)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account doesnt exist: "+form.Mail)

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountNotFound},
		})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account existence check failed - SQL query: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountBlocked},
		})
	}

	var step2code string
	var retry time.Duration
	if form.Process == OTPProcess {

		step2code, retry, err = OTPProcessor.Send(form.Mail, MailModels.SignInStarted, ctx.IP())
		if step2code == "" {
			Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 10_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
		}
	} else if form.Process == PasswordProcess {

		if !hasPassword {
			Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account doesnt have password")

			RequestProcessor.AddRateLimitWeight(ctx, 10_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.PasswordNotSet},
			})
		}
	} else {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Unknown process: "+form.Process)

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	token, err := TokenProcessor.CreateSignInToken(form, userID, form.Process, step2code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request Complete")

	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}
