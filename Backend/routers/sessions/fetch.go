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

// Fetch retrieves all active device sessions associated with the authenticated user.
//
// This function provides visibility into all sessions/devices currently associated
// with the user account. It:
//  1. Validates the access token and ensures it is still fresh.
//  2. Verifies that the current device/session is not revoked or blocked.
//  3. Fetches all device/session records for the user from the database.
//  4. Encrypts sensitive identifiers before returning them to the client.
//  5. Identifies the current session among all active sessions.
//
// Flow Summary:
//
//	validate access → verify device → fetch sessions → encrypt identifiers → build response → return
//
// Security Considerations:
// - Access token must be valid and unexpired.
// - Device access is verified to prevent revoked sessions from querying data.
// - UserID and DeviceID are encrypted before exposure to client.
// - Partial failures in row processing are tolerated (graceful degradation).
//
// Returns:
// - 200 OK with list of user sessions/devices
// - 401 Unauthorized (invalid/expired token or blocked device)
// - 500 Internal Server Error (database/encryption failures)
func Fetch(ctx fiber.Ctx) error {

	// Extract and validate access token
	access, err := TokenProcessor.ReadAccessToken(ctx)

	// Ensure token is valid and still fresh (not expired)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, fetchFilename, RequestProcessor.GetRequestId(ctx), "Requested by: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))

	// Ensure token is valid and still fresh (not expired)
	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	// Reject request if device/session is blocked
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, fetchFilename, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	// Fetch all device/session records for the user
	rows, err := Stores.SQLClient.Query(Config.CtxBG, "SELECT device_id, visits, remembered, created, updated, os, device, browser FROM devices WHERE user_id = $1", access.UserID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, fetchFilename, RequestProcessor.GetRequestId(ctx), "Device fetch failed - SQL query: "+err.Error())
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	defer rows.Close()

	// Prepare response container
	var response ResponseModels.UserDevicesResponseT

	// Iterate through all device records
	for rows.Next() {
		var deviceID int16
		var activity ResponseModels.SingleUserDeviceT

		// Map DB row into response struct
		err = rows.Scan(&deviceID, // device ID
			&activity.Count,      // visits count (refresh version)
			&activity.Remembered, // persistent session flag
			&activity.Created,    // session creation timestamp
			&activity.Updated,    // last refresh timestamp
			&activity.OS,         // parsed OS
			&activity.Device,     // device type
			&activity.Browser,    // browser info
		)
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Pack device data failed - SQL scan: "+err.Error())

			// Skip malformed rows but continue processing
			continue
		}

		// Encrypt device before exposing to client
		activity.ID, err = StringProcessor.EncryptInterfaceToB64(ResponseModels.SingleDeviceT{
			UserID:   access.UserID,
			DeviceID: deviceID,
		})
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Device encrypt failed: "+err.Error())

			continue
		}

		// Identify current session (used for refresh tracking on client)
		if deviceID == access.DeviceID {
			response.Current = activity.ID
		}

		// Append activity record to response
		response.Devices = append(response.Devices, activity)
	}

	// Loop ended prematurely
	if err = rows.Err(); err != nil {
		Logs.RootLogger.Add(Logs.Warn, fetchFilename, RequestProcessor.GetRequestId(ctx), "Loop ended prematurely: "+err.Error())
	}

	Logs.RootLogger.Add(Logs.Info, fetchFilename, RequestProcessor.GetRequestId(ctx), "Completed request")
	// Return aggregated session/activity data
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   response,
	})
}
