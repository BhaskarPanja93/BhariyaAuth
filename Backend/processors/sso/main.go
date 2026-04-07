package sso

import (
	HTMLTemplates "BhariyaAuth/models/html"
	"time"

	"github.com/gofiber/fiber/v3"
)

// SuccessPopup returns an HTML response indicating successful SSO authentication.
//
// This function is used at the end of the SSO flow (OAuth callback).
// It renders an HTML page (typically inside a popup window) that:
//  1. Contains the authentication token.
//  2. Includes token expiry information.
//  3. Allows the frontend (parent window) to retrieve the token.
//
// Typical Usage:
//   - Opened as a popup during SSO signin.
//   - On success, this HTML communicates back to the main application
//     (e.g., via window.postMessage or window.opener).
//
// Parameters:
// - token: authentication token generated after successful SSO.
// - expires: expiration time of the token.
//
// Response:
// - Content-Type: text/html
// - Body: HTML page containing token + expiry metadata.
//
// Security Considerations:
// - Token is embedded in HTML → must ensure:
//   - HTTPS is always used.
//   - No caching of response.
//   - Controlled origin communication in frontend.
//
// - Expiry is formatted in RFC3339 for consistent parsing.
//
// Returns:
// - HTML response to client.
func SuccessPopup(ctx fiber.Ctx, token string, expires time.Time) error {

	return ctx.
		Type("html").
		SendString(
			HTMLTemplates.SSOSuccessPopup(
				token,
				expires.Format(time.RFC3339),
			),
		)
}

// FailurePopup returns an HTML response indicating failed SSO authentication.
//
// This function renders an HTML page in case of SSO failure.
// It communicates the failure reason back to the user or frontend.
//
// Typical Usage:
// - Displayed in popup window when SSO flow fails.
// - Can trigger UI updates in parent window.
//
// Parameters:
// - reason: human-readable failure reason.
//
// Response:
// - Content-Type: text/html
// - Body: HTML page with error message.
//
// Security Considerations:
// - Ensure failure reasons do not leak sensitive internal details.
// - Avoid exposing provider-specific errors directly.
//
// Returns:
// - HTML response to client.
func FailurePopup(ctx fiber.Ctx, reason string) error {

	return ctx.
		Type("html").
		SendString(
			HTMLTemplates.SSOFailurePopup(reason),
		)
}
