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

// Logout terminates the current authenticated session by revoking the device session
// and clearing all authentication-related cookies.
//
// This function handles user logout by:
//  1. Validating the refresh token and CSRF token.
//  2. Removing all authentication-related cookies from the client.
//  3. Deleting the corresponding device/session from the database.
//
// Flow Summary:
//
//	validate refresh → verify CSRF → clear cookies → delete session → return success
//
// Security Considerations:
// - CSRF validation ensures logout requests are legitimate.
// - Session is deleted server-side to prevent reuse of refresh tokens.
// - All cookies (auth, MFA, SSO) are cleared to fully invalidate client state.
// - Even if session deletion fails silently, cookie removal prevents further use.
//
// Dependencies:
// - Requires valid refresh token (cookie-based).
// - Requires CSRF token validation.
//
// Returns:
// - 200 OK on successful logout
// - 422 Unprocessable Entity if token is invalid or missing
func Logout(ctx fiber.Ctx) error {

	// Extract refresh token from request
	access, err := TokenProcessor.ReadAccessToken(ctx)

	// Validate token presence and CSRF protection
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, logoutFileName, RequestProcessor.GetRequestId(ctx), "Access invalid/expired")

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	Logs.RootLogger.Add(Logs.Error, logoutFileName, RequestProcessor.GetRequestId(ctx), "Request for: "+strconv.Itoa(int(access.UserID))+" "+strconv.Itoa(int(access.DeviceID)))

	// Clear all authentication-related cookies from client
	// This ensures client cannot continue using stale tokens
	CookieProcessor.DetachAuthCookies(ctx)
	CookieProcessor.DetachMFACookies(ctx)
	CookieProcessor.DetachSSOCookies(ctx)

	// Delete the session/device entry from database
	// This revokes the refresh token server-side
	err = AccountProcessor.DenySingleDeviceFromRenewing(access.UserID, access.DeviceID)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, logoutFileName, RequestProcessor.GetRequestId(ctx), "Device revoke failed: "+err.Error())

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}

	Logs.RootLogger.Add(Logs.Error, logoutFileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return success response indicating auth state has changed
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
		})
}
