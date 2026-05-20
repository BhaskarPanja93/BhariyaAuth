package logs

import (
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"

	"github.com/gofiber/fiber/v3"
)

const availableFilename = "routers/logs/available"

func Available(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessToken(ctx)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) || UserTypes.Find(access.UserType).Authority < UserTypes.All.Moderator.Authority {
		Logs.RootLogger.Add(Logs.Blocked, availableFilename, RequestProcessor.GetRequestId(ctx), "Access invalid/expired/lacks permissions")

		RequestProcessor.AddRateLimitWeight(ctx, 120_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, availableFilename, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, availableFilename, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   Logs.RootLogger.CheckAvailableLogFiles(),
	})

}
