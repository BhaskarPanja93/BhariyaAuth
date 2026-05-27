package sessions

import (
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const revokeFileName = "routers/sessions/revoke"

func Revoke(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessHeader(ctx)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))

	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	mfa, err := TokenProcessor.ReadMFAHeader(ctx)
	if err != nil || !mfa.Verified || mfa.UserID != access.UserID || mfa.DeviceID != access.DeviceID {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "MFA invalid/missing")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusForbidden)
	}

	form := new(FormModels.DeviceRevokeForm)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Invalid form")

		RequestProcessor.AddRateLimitWeight(ctx, 20_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	var device ResponseModels.SingleDeviceT
	err = StringProcessor.DecryptInterfaceFromB64(form.Device, &device)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Device decrypt failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if form.All == "yes" {
		Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for"+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID))+" revoke all devices")

		err = AccountProcessor.DenyAllDevicesFromRenewing(access.UserID)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Revoke failed: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 60_000)

			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
		Logs.RootLogger.Add(Logs.Info, revokeFileName, RequestProcessor.GetRequestId(ctx), "Request Complete")
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success: true,
		})
	}

	if access.UserID != device.UserID {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Data belongs to different user")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for"+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID))+" revoke: "+strconv.Itoa(int(device.DeviceID)))

	err = AccountProcessor.DenySingleDeviceFromRenewing(device.UserID, device.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Revoke failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	Logs.RootLogger.Add(Logs.Info, revokeFileName, RequestProcessor.GetRequestId(ctx), "Request Complete")
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
