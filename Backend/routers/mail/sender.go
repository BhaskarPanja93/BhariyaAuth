package mail

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	MailNotifier "BhariyaAuth/processors/mail"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const senderFilename = "routers/mail/sender"

func Sender(ctx fiber.Ctx) error {
	access, err := TokenProcessor.ReadAccessHeader(ctx)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) || UserTypes.Find(access.UserType).Authority < UserTypes.All.Admin.Authority {
		Logs.RootLogger.Add(Logs.Blocked, senderFilename, RequestProcessor.GetRequestId(ctx), "Access invalid/expired/lacks permissions")

		RequestProcessor.AddRateLimitWeight(ctx, 120_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, senderFilename, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, senderFilename, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 120_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	form := new(FormModels.MailSendForm)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, senderFilename, RequestProcessor.GetRequestId(ctx), "Invalid form")
		RequestProcessor.AddRateLimitWeight(ctx, 120_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	var mails []string

	if form.Audience == "individuals" {
		for _, mail := range form.Recipients {
			mails = append(mails, mail)
		}
	} else if form.Audience == "groups" || form.Audience == "everyone" {
		var rows pgx.Rows
		if form.Audience == "groups" {
			rows, err = Stores.SQLClient.Query(Config.CtxBG, "SELECT mail FROM users WHERE type = ANY($1)", form.Recipients)
		} else {
			rows, err = Stores.SQLClient.Query(Config.CtxBG, "SELECT mail FROM users")
		}
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, senderFilename, RequestProcessor.GetRequestId(ctx), "Mail list fetch failed: "+err.Error())
			RequestProcessor.AddRateLimitWeight(ctx, 10_000)

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
		}

		defer rows.Close()
		var mail string

		for rows.Next() {
			err = rows.Scan(&mail)
			if err != nil {
				Logs.RootLogger.Add(Logs.Error, senderFilename, RequestProcessor.GetRequestId(ctx), "Row Scan error: "+err.Error())
				RequestProcessor.AddRateLimitWeight(ctx, 10_000)

				return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
					Notifications: []string{"Action halted"},
				})
			}
			mails = append(mails, mail)
		}
		err = rows.Err()
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, senderFilename, RequestProcessor.GetRequestId(ctx), "Query unknown error: "+err.Error())
			RequestProcessor.AddRateLimitWeight(ctx, 10_000)

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{"Action halted"},
			})
		}
	}

	err = MailNotifier.Raw(mails, form.Subject, form.Body)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, senderFilename, RequestProcessor.GetRequestId(ctx), "Mail sender error: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{"Mail sender returned error"},
		})
	}

	Logs.RootLogger.Add(Logs.Info, senderFilename, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Notifications: []string{
			"Mail send queued for " + strconv.Itoa(len(mails)) + " users",
			"NOTE: mails sent to 500+ clients might fail silently"},
	})
}
