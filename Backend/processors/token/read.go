package token

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	StringProcessor "BhariyaAuth/processors/string"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func readMFAValue(ctx fiber.Ctx) string {
	return strings.TrimSpace(ctx.Cookies(Config.MFATokenInHeader))
}

func readBearerValue(ctx fiber.Ctx) string {
	header := strings.TrimSpace(ctx.Get(Config.AccessTokenInHeader))
	if len(header) < 7 {
		return header
	}

	if strings.EqualFold(header[:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}

	return header
}

func ReadMFAHeader(ctx fiber.Ctx) (TokenModels.MFAToken, error) {
	var mfa TokenModels.MFAToken
	token := readMFAValue(ctx)

	err := StringProcessor.DecryptInterfaceFromB64(token, &mfa)
	if err != nil {
		return mfa, err
	}

	if mfa.TokenType != mfaTokenType {
		return mfa, errors.New("read mfa token: incorrect type")
	}

	return mfa, nil
}

func ReadAccessHeader(ctx fiber.Ctx) (TokenModels.AccessToken, error) {
	var access TokenModels.AccessToken
	token := readBearerValue(ctx)

	err := StringProcessor.DecryptInterfaceFromB64(token, &access)
	if err != nil {
		return access, err
	}

	if access.TokenType != accessTokenType {
		return access, errors.New("read access token: incorrect type")
	}

	return access, nil
}

func ReadHeaderCSRF(ctx fiber.Ctx) string {
	return strings.TrimSpace(ctx.Get(Config.CSRFInHeader))
}

func ReadRefreshCookie(ctx fiber.Ctx) (TokenModels.RefreshToken, error) {
	var refresh TokenModels.RefreshToken
	cookie := strings.TrimSpace(ctx.Cookies(Config.RefreshTokenInCookie))

	err := StringProcessor.DecryptInterfaceFromB64(cookie, &refresh)
	if err != nil {
		return refresh, err
	}

	if refresh.TokenType != refreshTokenType {
		return refresh, errors.New("read refresh token: incorrect type")
	}

	return refresh, nil
}

func ReadMFAToken(token string) (TokenModels.MFAToken, error) {
	var mfa TokenModels.MFAToken

	err := StringProcessor.DecryptInterfaceFromB64(strings.TrimSpace(token), &mfa)
	if err != nil {
		return mfa, err
	}

	if mfa.TokenType != mfaTokenType {
		return mfa, errors.New("read mfa token: incorrect type")
	}

	return mfa, nil
}

func ReadSSOToken(token string) (TokenModels.SSOState, error) {
	var sso TokenModels.SSOState

	err := StringProcessor.DecryptInterfaceFromB64(strings.TrimSpace(token), &sso)
	if err != nil {
		return sso, err
	}

	if sso.TokenType != ssoTokenType {
		return sso, errors.New("read sso token: incorrect type")
	}

	return sso, nil
}

func ReadPasswordResetToken(token string) (TokenModels.PasswordReset, error) {
	var reset TokenModels.PasswordReset

	err := StringProcessor.DecryptInterfaceFromB64(strings.TrimSpace(token), &reset)
	if err != nil {
		return reset, err
	}

	if reset.TokenType != passwordResetTokenType {
		return reset, errors.New("read password reset token: incorrect type")
	}

	return reset, nil
}

func ReadSignInToken(token string) (TokenModels.SignIn, error) {
	var signIn TokenModels.SignIn

	err := StringProcessor.DecryptInterfaceFromB64(strings.TrimSpace(token), &signIn)
	if err != nil {
		return signIn, err
	}

	if signIn.TokenType != signInTokenType {
		return signIn, errors.New("read signin token: incorrect type")
	}

	return signIn, nil
}

func ReadSignUpToken(token string) (TokenModels.SignUp, error) {
	var signUp TokenModels.SignUp

	err := StringProcessor.DecryptInterfaceFromB64(strings.TrimSpace(token), &signUp)
	if err != nil {
		return signUp, err
	}

	if signUp.TokenType != signUpTokenType {
		return signUp, errors.New("read signup token: incorrect type")
	}

	return signUp, nil
}
