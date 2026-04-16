package token

import (
	TokenModels "BhariyaAuth/models/tokens"
	RequestProcessor "BhariyaAuth/processors/request"

	"github.com/gofiber/fiber/v3"
)

// VerifyCSRF validates CSRF token using double-submit cookie pattern.
//
// Overview:
// This function ensures that the CSRF token:
//  1. Exists in the refresh token.
//  2. Matches the CSRF cookie.
//  3. Matches the CSRF header.
//
// Flow:
//
//	refresh token CSRF → compare with cookie → compare with header
//
// Security:
// - Prevents CSRF attacks by requiring:
//   - attacker cannot read cookie
//   - attacker cannot forge matching header
//
// Returns:
// - true if CSRF is valid.
// - false otherwise.
func VerifyCSRF(ctx fiber.Ctx, refresh TokenModels.RefreshToken) bool {

	// CSRF must exist
	if refresh.CSRF == "" {
		return false
	}

	cookie := ReadCookieCSRF(ctx)
	header := ReadHeaderCSRF(ctx)

	// All values must match
	return cookie != "" &&
		header != "" &&
		cookie == refresh.CSRF &&
		header == refresh.CSRF
}

// AccessIsFresh checks whether access token is still valid (not expired).
//
// Logic:
//
//	current time < expiry → valid
//
// Returns:
// - true if token is still valid.
// - false if expired.
func AccessIsFresh(ctx fiber.Ctx, access TokenModels.AccessToken) bool {

	return RequestProcessor.GetRequestTime(ctx).Before(access.Expiry)
}

// RefreshIsFresh checks whether refresh token is still valid (not expired).
//
// Logic:
//
//	current time < expiry → valid
//
// Returns:
// - true if token is still valid.
// - false if expired.
func RefreshIsFresh(ctx fiber.Ctx, refresh TokenModels.RefreshToken) bool {

	return RequestProcessor.GetRequestTime(ctx).Before(refresh.Expiry)
}
