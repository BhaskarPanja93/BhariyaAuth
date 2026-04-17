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

// Revoke invalidates one or more user sessions (devices) by preventing
// their refresh tokens from being used for renewal.
//
// This function allows an authenticated user to:
//  1. Revoke a specific device/session.
//  2. Revoke all sessions across all devices.
//
// It ensures that:
//   - The request is authenticated and tied to a valid session.
//   - The requesting user can only revoke their own sessions.
//   - Device identifiers are securely handled (encrypted externally).
//   - Revoked sessions cannot obtain new access tokens (refresh denied).
//
// Flow Summary:
//
//	validate access → verify device → parse request → validate ownership → (revoke all | revoke one) → return success
//
// Security Considerations:
// - Access token must be valid and fresh.
// - Device-level access must not be blocked.
// - UserID and DeviceID are encrypted externally and decrypted server-side.
// - Ownership check prevents cross-user session revocation.
// - Rate limiting protects against abuse (mass revocation attempts).
//
// Request Body:
// - user (string) → encrypted user ID
// - refresh (string) → encrypted device ID (optional if `all=yes`)
// - all ("yes" | "no") → whether to revoke all sessions
//
// Returns:
// - 200 OK on success
// - 401 Unauthorized (invalid session or ownership mismatch)
// - 422 Unprocessable Entity (invalid input)
// - 500 Internal Server Error (system failures)
func Revoke(ctx fiber.Ctx) error {

	// Extract and validate access token
	access, err := TokenProcessor.ReadAccessToken(ctx)
	if err != nil || !TokenProcessor.AccessIsFresh(ctx, access) {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(access.UserID))+strconv.Itoa(int(access.DeviceID)))

	// Ensure current device/session is not blocked
	revoked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access revoke check failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if revoked {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Access revoked")
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse request payload
	form := new(FormModels.DeviceRevokeForm)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Invalid form")

		RequestProcessor.AddRateLimitWeight(ctx, 20_000) // 30 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	var device ResponseModels.SingleDeviceT
	// Decrypt user identifier from request
	err = StringProcessor.DecryptInterfaceFromB64(form.Device, &device)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Device decrypt failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if form.All == "yes" {
		Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for"+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID))+" revoke all devices")

		// Prevent all devices from renewing (global logout)
		err = AccountProcessor.DenyAllDevicesFromRenewing(access.UserID)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Revoke failed: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
		Logs.RootLogger.Add(Logs.Info, revokeFileName, RequestProcessor.GetRequestId(ctx), "Request Complete")
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success: true,
		})
	}

	// Ensure user can only revoke their own sessions
	if access.UserID != device.UserID {
		Logs.RootLogger.Add(Logs.Blocked, revokeFileName, RequestProcessor.GetRequestId(ctx), "Data belongs to different user")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	Logs.RootLogger.Add(Logs.Intent, revokeFileName, RequestProcessor.GetRequestId(ctx), "Requested for"+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID))+" revoke: "+strconv.Itoa(int(device.DeviceID)))

	// Revoke specific device/session
	err = AccountProcessor.DenySingleDeviceFromRenewing(device.UserID, device.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, revokeFileName, RequestProcessor.GetRequestId(ctx), "Revoke failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	Logs.RootLogger.Add(Logs.Info, revokeFileName, RequestProcessor.GetRequestId(ctx), "Request Complete")
	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
