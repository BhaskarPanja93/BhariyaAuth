package sessions

import (
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	FormProcessor "BhariyaAuth/processors/form"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"

	"github.com/gofiber/fiber/v3"
)

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
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	// Ensure current device/session is not blocked
	blocked, err := AccountProcessor.CheckDeviceAccessDenied(access.UserID, access.DeviceID)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}
	if blocked {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse request payload
	form := new(FormModels.DeviceRevokeForm)
	if !FormProcessor.ReadFormData(ctx, form) {
		RateLimitProcessor.Add(ctx, 20_000) // 30 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Decrypt user identifier from request
	userID, err := StringProcessor.DecryptUserID(form.User)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Ensure user can only revoke their own sessions
	if access.UserID != userID {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if form.All == "yes" {

		// Prevent all devices from renewing (global logout)
		err = AccountProcessor.DenyAllDevicesFromRenewing(userID)
		if err != nil {
			RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success: true,
			})
	}

	// Decrypt device identifier
	deviceID, err := StringProcessor.DecryptDeviceID(form.Refresh)
	if err != nil {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Revoke specific device/session
	err = AccountProcessor.DenySingleDeviceFromRenewing(userID, deviceID)
	if err != nil {
		RateLimitProcessor.Add(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
		})
}
