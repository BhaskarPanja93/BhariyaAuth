package access

import (
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const logoutFileName = "routers/access/logout"

func Logout(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessToken(ctx)

	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, logoutFileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Intent, logoutFileName, RequestProcessor.GetRequestId(ctx), "Request for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))

	CookieProcessor.DetachAuthCookies(ctx)
	CookieProcessor.DetachMFACookies(ctx)
	CookieProcessor.DetachSSOCookies(ctx)

	err = AccountProcessor.DenySingleDeviceFromRenewing(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, logoutFileName, RequestProcessor.GetRequestId(ctx), "Device revoke failed: "+err.Error())

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}

	Logs.RootLogger.Add(Logs.Info, logoutFileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
		})
}
