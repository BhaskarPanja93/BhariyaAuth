package logs

import (
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const dateFileName = "routers/logs/date"

func Date(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessHeader(ctx)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) || UserTypes.Find(access.UserType).Authority < UserTypes.All.Moderator.Authority {
		Logs.RootLogger.Add(Logs.Blocked, dateFileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired/lacks permissions")

		RequestProcessor.AddRateLimitWeight(ctx, 120_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, dateFileName, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, dateFileName, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	date := ctx.Params(DateParam)
	Logs.RootLogger.Add(Logs.Intent, dateFileName, RequestProcessor.GetRequestId(ctx), "Requested by: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID))+" for date "+date)

	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   Logs.RootLogger.ReadLogFile(date),
	})
}
