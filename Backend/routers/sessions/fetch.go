package sessions

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const fetchFilename = "routers/sessions/fetch"

func Fetch(ctx fiber.Ctx) error {

	access, err := TokenProcessor.ReadAccessToken(ctx)

	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, fetchFilename, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))

	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	rows, err := Stores.SQLClient.Query(Config.CtxBG, "SELECT device_id, visits, remembered, created, updated, os, device, browser FROM devices WHERE user_id = $1", access.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, fetchFilename, RequestProcessor.GetRequestId(ctx), "Device fetch failed - SQL query: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	defer rows.Close()

	var response ResponseModels.UserDevicesResponseT

	for rows.Next() {
		var deviceID int16
		var activity ResponseModels.SingleUserDeviceT

		err = rows.Scan(&deviceID,
			&activity.Count,
			&activity.Remembered,
			&activity.Created,
			&activity.Updated,
			&activity.OS,
			&activity.Device,
			&activity.Browser,
		)
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Pack device data failed - SQL scan: "+err.Error())

			continue
		}

		activity.ID, err = StringProcessor.EncryptInterfaceToB64(ResponseModels.SingleDeviceT{
			UserID:   access.UserID,
			DeviceID: deviceID,
		})
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Device encrypt failed: "+err.Error())

			continue
		}

		if deviceID == access.DeviceID {
			response.Current = activity.ID
		}

		response.Devices = append(response.Devices, activity)
	}

	if err = rows.Err(); err != nil {
		Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Loop ended prematurely: "+err.Error())
	}

	Logs.RootLogger.Add(Logs.Info, fetchFilename, RequestProcessor.GetRequestId(ctx), "Request Complete")
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   response,
	})
}
