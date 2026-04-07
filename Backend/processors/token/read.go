package token

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	StringProcessor "BhariyaAuth/processors/string"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ReadHeaderCSRF extracts CSRF token from request header.
//
// Source:
// - Header name defined in Config.CSRFInHeader.
//
// Usage:
// - Used in double-submit CSRF protection.
//
// Returns:
// - CSRF token string (empty if missing).
func ReadHeaderCSRF(ctx fiber.Ctx) string {
	return ctx.Get(Config.CSRFInHeader)
}

// ReadCookieCSRF extracts CSRF token from cookie.
//
// Source:
// - Cookie name defined in Config.CSRFInCookie.
//
// Usage:
// - Compared against header CSRF for validation.
func ReadCookieCSRF(ctx fiber.Ctx) string {
	return ctx.Cookies(Config.CSRFInCookie)
}

// ReadAccessToken extracts, decrypts, and validates access token from header.
//
// Flow:
//
//	read header → strip "Bearer " → decrypt → validate type
//
// Returns:
// - AccessToken struct if valid.
// - error if:
//   - missing/invalid token
//   - decryption fails
//   - token type mismatch.
func ReadAccessToken(ctx fiber.Ctx) (TokenModels.AccessToken, error) {

	var access TokenModels.AccessToken

	// Extract token from Authorization header
	header := strings.TrimPrefix(ctx.Get(Config.AccessTokenInHeader), "Bearer ")

	// Decrypt token
	err := StringProcessor.DecryptInterfaceFromString(header, &access)
	if err != nil {
		return access, err
	}

	// Validate token type
	if access.TokenType != accessTokenType {
		return access, IncorrectAccessTokenTypeError
	}

	return access, nil
}

// ReadRefreshToken extracts, decrypts, and validates refresh token from cookie.
//
// Flow:
//
//	read cookie → decrypt → validate type
func ReadRefreshToken(ctx fiber.Ctx) (TokenModels.RefreshToken, error) {

	var refresh TokenModels.RefreshToken

	cookie := ctx.Cookies(Config.RefreshTokenInCookie)

	err := StringProcessor.DecryptInterfaceFromString(cookie, &refresh)
	if err != nil {
		return refresh, err
	}

	if refresh.TokenType != refreshTokenType {
		return refresh, IncorrectRefreshTokenTypeError
	}

	return refresh, nil
}

// ReadMFAToken decrypts and validates MFA token.
func ReadMFAToken(token string) (TokenModels.MFAToken, error) {

	var mfa TokenModels.MFAToken

	err := StringProcessor.DecryptInterfaceFromString(token, &mfa)
	if err != nil {
		return mfa, err
	}

	if mfa.TokenType != mfaTokenType {
		return mfa, IncorrectMFATokenTypeError
	}

	return mfa, nil
}

// ReadSSOToken decrypts and validates SSO state token.
func ReadSSOToken(token string) (TokenModels.SSOState, error) {

	var sso TokenModels.SSOState

	err := StringProcessor.DecryptInterfaceFromString(token, &sso)
	if err != nil {
		return sso, err
	}

	if sso.TokenType != ssoTokenType {
		return sso, IncorrectSSOTokenTypeError
	}

	return sso, nil
}

// ReadPasswordResetToken decrypts and validates password reset token.
func ReadPasswordResetToken(token string) (TokenModels.PasswordReset, error) {

	var reset TokenModels.PasswordReset

	err := StringProcessor.DecryptInterfaceFromString(token, &reset)
	if err != nil {
		return reset, err
	}

	if reset.TokenType != passwordResetTokenType {
		return reset, IncorrectPasswordResetTokenTypeError
	}

	return reset, nil
}

// ReadSignInToken decrypts and validates signin step token.
func ReadSignInToken(token string) (TokenModels.SignIn, error) {

	var signIn TokenModels.SignIn

	err := StringProcessor.DecryptInterfaceFromString(token, &signIn)
	if err != nil {
		return signIn, err
	}

	if signIn.TokenType != signInTokenType {
		return signIn, IncorrectSignInTokenTypeError
	}

	return signIn, nil
}

// ReadSignUpToken decrypts and validates registration step token.
func ReadSignUpToken(token string) (TokenModels.SignUp, error) {

	var signUp TokenModels.SignUp

	err := StringProcessor.DecryptInterfaceFromString(token, &signUp)
	if err != nil {
		return signUp, err
	}

	if signUp.TokenType != signUpTokenType {
		return signUp, IncorrectSignUpTokenTypeError
	}

	return signUp, nil
}
