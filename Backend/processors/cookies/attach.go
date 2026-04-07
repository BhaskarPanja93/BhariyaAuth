package cookies

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	"time"

	"github.com/gofiber/fiber/v3"
)

// AttachAuthCookies sets authentication-related cookies (Refresh Token + CSRF)
// on the response. These cookies are used for session persistence and request validation.
func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenCombined) {

	start := ctx.Locals("request-start").(time.Time)

	Refresh := fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,    // Cookie name from configuration
		Value:    token.RefreshToken,             // Actual refresh token value
		HTTPOnly: true,                           // Prevent JavaScript access (mitigates XSS attacks)
		Secure:   true,                           // Ensures cookie is only sent over HTTPS
		SameSite: fiber.CookieSameSiteStrictMode, // Prevents sending cookie in cross-site requests (CSRF protection)
		Domain:   Config.Domain,                  // Restricts cookie to a specific domain
		Path:     "/",                            // Specific path
	}

	// CSRF token cookie: used for validating state-changing requests.
	Csrf := fiber.Cookie{
		Name:     Config.CSRFInCookie,            // Cookie name for CSRF token
		Value:    token.CSRF,                     // CSRF token value
		HTTPOnly: false,                          // Accessible via JavaScript (required for client-side inclusion in headers)
		Secure:   true,                           // Only sent over HTTPS
		SameSite: fiber.CookieSameSiteStrictMode, // Strict CSRF protection
		Domain:   Config.Domain,                  // Domain restriction
		Path:     "/",                            // Specific path
	}

	// If "Remember Me" is enabled, persist cookies for a longer duration
	if token.RememberMe {
		// Set cookie expiration time in seconds
		Refresh.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Csrf.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Refresh.Expires = start.Add(Config.RefreshTokenExpireDelta)
		Csrf.Expires = start.Add(Config.RefreshTokenExpireDelta)
	} else {
		// Session cookies (expire when browser closes)
		Refresh.MaxAge = 0
		Csrf.MaxAge = 0
	}

	// Attach cookies to HTTP response
	ctx.Cookie(&Refresh)
	ctx.Cookie(&Csrf)
}

// AttachSSOCookie sets a temporary cookie used during SSO (Single Sign-On) flows.
// Typically used to maintain state between identity provider redirects.
func AttachSSOCookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.SSOStateInCookie,                     // Cookie name for SSO state tracking
		Value:    value,                                       // State value (usually random nonce)
		MaxAge:   int(Config.SSOCookieExpireDelta.Seconds()),  // Expiration duration
		Expires:  time.Now().Add(Config.SSOCookieExpireDelta), // Expiry
		Secure:   true,                                        // HTTPS only
		HTTPOnly: true,                                        // Not accessible via JavaScript
		SameSite: fiber.CookieSameSiteLaxMode,                 // Allows cookie on top-level navigation (required for OAuth redirects)
		Domain:   Config.Domain,                               // Domain restriction
		Path:     "/",                                         // Specific path
	})
}

// AttachMFACookie sets a cookie used for Multi-Factor Authentication (MFA) state.
// This is typically short-lived and used during verification flows.
func AttachMFACookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,        // Cookie name for MFA token
		Value:    value,                          // MFA token or session identifier
		MaxAge:   0,                              // Delete after browser close
		Secure:   true,                           // HTTPS only
		HTTPOnly: true,                           // Not accessible via JavaScript
		SameSite: fiber.CookieSameSiteStrictMode, // Prevents cross-site usage
		Domain:   Config.Domain,                  // Domain restriction
		Path:     "/",                            // Specific path
	})
}
