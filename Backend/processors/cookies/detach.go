package cookies

import (
	Config "BhariyaAuth/constants/config"

	"github.com/gofiber/fiber/v3"
)

// DetachAuthCookies clears authentication cookies by overwriting them with empty values.
// This effectively logs the user out from a cookie/session perspective.
func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.RefreshTokenInCookie)
	ctx.ClearCookie(Config.CSRFInCookie)
}

// DetachSSOCookies clears the SSO state cookie.
// Used after completing or aborting an SSO flow.
func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.SSOStateInCookie)
}

// DetachMFACookies clears the MFA token cookie.
// Typically called after MFA verification or timeout.
func DetachMFACookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.MFATokenInCookie)
}
