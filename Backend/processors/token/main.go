package token

import (
	Config "BhariyaAuth/constants/config"
	FormModels "BhariyaAuth/models/requests"
	TokenModels "BhariyaAuth/models/tokens"
	StringProcessor "BhariyaAuth/processors/string"
	"math"
	"time"

	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func Encrypt(v interface{}) (string, bool) {
	marshalled, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	encrypted, ok := StringProcessor.Encrypt(marshalled)
	if !ok {
		return "", false
	}
	return encrypted, true
}

func CreateFreshToken(ctx fiber.Ctx, userID int32, deviceID int16, userType string, remember bool, identifierType string) (TokenModels.NewTokenCombinedT, bool) {
	now := ctx.Locals("request-start").(time.Time)
	accessExpires := now.Add(Config.AccessTokenExpireDelta)
	refreshExpires := now.Add(Config.RefreshTokenExpireDelta)
	csrf := StringProcessor.SafeString(128)
	atEnc, ok := Encrypt(TokenModels.AccessTokenT{
		TokenType:    accessTokenType,
		UserID:       userID,
		DeviceID:     deviceID,
		UserType:     userType,
		AccessExpiry: accessExpires,
		RememberMe:   remember,
	})
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtEnc, ok := Encrypt(TokenModels.RefreshTokenT{
		TokenType:      refreshTokenType,
		UserID:         userID,
		DeviceID:       deviceID,
		Visits:         math.MinInt16,
		Created:        now,
		Updated:        now,
		Expiry:         refreshExpires,
		UserType:       userType,
		CSRF:           csrf,
		RememberMe:     remember,
		IdentifierType: identifierType,
	})
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	return TokenModels.NewTokenCombinedT{
		AccessToken:   atEnc,
		RefreshToken:  rtEnc,
		AccessExpires: accessExpires,
		CSRF:          csrf,
		RememberMe:    remember,
	}, true
}

func CreateRenewToken(ctx fiber.Ctx, refresh *TokenModels.RefreshTokenT) (*TokenModels.NewTokenCombinedT, bool) {
	var combined TokenModels.NewTokenCombinedT
	now := ctx.Locals("request-start").(time.Time)
	accessExpires := now.Add(Config.AccessTokenExpireDelta)
	refreshExpires := now.Add(Config.RefreshTokenExpireDelta)
	csrf := StringProcessor.SafeString(128)
	atEnc, ok := Encrypt(TokenModels.AccessTokenT{
		TokenType:    accessTokenType,
		UserID:       refresh.UserID,
		DeviceID:     refresh.DeviceID,
		UserType:     refresh.UserType,
		AccessExpiry: accessExpires,
		RememberMe:   refresh.RememberMe,
	})
	if !ok {
		return &combined, false
	}
	refresh.CSRF = csrf
	refresh.Expiry = refreshExpires
	if refresh.Visits >= math.MaxInt16 {
		refresh.Visits = math.MinInt16
	} else {
		refresh.Visits++
	}
	rtEnc, ok := Encrypt(refresh)
	if !ok {
		return &combined, false
	}
	combined.AccessToken = atEnc
	combined.RefreshToken = rtEnc
	combined.CSRF = csrf
	combined.RememberMe = refresh.RememberMe
	return &combined, true
}

func ReadAccessToken(ctx fiber.Ctx) (TokenModels.AccessTokenT, bool) {
	var access TokenModels.AccessTokenT
	header := strings.TrimPrefix(ctx.Get(Config.AccessTokenInHeader), "Bearer ")
	tokenDec, ok := StringProcessor.Decrypt(header)
	return access, ok && json.Unmarshal(tokenDec, &access) == nil && access.TokenType == accessTokenType
}

func ReadRefreshToken(ctx fiber.Ctx) (*TokenModels.RefreshTokenT, bool) {
	var refresh TokenModels.RefreshTokenT
	cookie := ctx.Cookies(Config.RefreshTokenInCookie)
	tokenDec, ok := StringProcessor.Decrypt(cookie)
	return &refresh, ok && json.Unmarshal(tokenDec, &refresh) == nil && refresh.TokenType == refreshTokenType
}

func MatchCSRF(ctx fiber.Ctx, refresh *TokenModels.RefreshTokenT) bool {
	cookie := ctx.Cookies(Config.CSRFInCookie)
	header := ctx.Get(Config.CSRFInHeader)
	return refresh.CSRF == header && refresh.CSRF == cookie
}

func CreateMFAToken(ctx fiber.Ctx, userID int32, deviceID int16, step2code string) (string, bool) {
	return Encrypt(TokenModels.MFATokenT{
		TokenType: mfaTokenType,
		Step2Code: step2code,
		UserID:    userID,
		DeviceID:  deviceID,
		Created:   ctx.Locals("request-start").(time.Time),
		Verified:  true,
	})
}

func ReadMFAToken(token string) (*TokenModels.MFATokenT, bool) {
	var mfa TokenModels.MFATokenT
	tokenDec, ok := StringProcessor.Decrypt(token)
	return &mfa, ok && json.Unmarshal(tokenDec, &mfa) == nil && mfa.TokenType == mfaTokenType
}

func CreateSSOToken(ctx fiber.Ctx, provider string) (string, bool) {
	return Encrypt(TokenModels.SSOStateT{
		TokenType:  ssoTokenType,
		Provider:   provider,
		Expiry:     ctx.Locals("request-start").(time.Time).Add(Config.SSOCookieExpireDelta),
		RememberMe: ctx.Query("remember", "no") == "yes",
	})
}

func ReadSSOToken(state string) (*TokenModels.SSOStateT, bool) {
	var sso TokenModels.SSOStateT
	tokenDec, ok := StringProcessor.Decrypt(state)
	return &sso, ok && json.Unmarshal(tokenDec, &sso) == nil && sso.TokenType == ssoTokenType
}

func CreatePasswordResetToken(mail string, userID int32, step2code string) (string, bool) {
	return Encrypt(TokenModels.PasswordResetT{
		TokenType:   passwordResetTokenType,
		MailAddress: mail,
		UserID:      userID,
		Step2Code:   step2code,
	})
}

func ReadPasswordResetToken(token string) (*TokenModels.PasswordResetT, bool) {
	var passwordReset TokenModels.PasswordResetT
	tokenDec, ok := StringProcessor.Decrypt(token)
	return &passwordReset, ok && json.Unmarshal(tokenDec, &passwordReset) == nil && passwordReset.TokenType == passwordResetTokenType
}

func CreateSignInToken(form *FormModels.SignInForm1, userID int32, step2process string, step2code string) (string, bool) {
	return Encrypt(TokenModels.SignInT{
		TokenType:    signInTokenType,
		UserID:       userID,
		RememberMe:   form.Remember == "yes",
		Step2Process: step2process,
		Step2Code:    step2code,
		MailAddress:  form.Mail,
	})
}

func ReadSignInToken(token string) (*TokenModels.SignInT, bool) {
	var signIn TokenModels.SignInT
	tokenDec, ok := StringProcessor.Decrypt(token)
	return &signIn, ok && json.Unmarshal(tokenDec, &signIn) == nil && signIn.TokenType == signInTokenType
}

func CreateSignUpToken(form *FormModels.RegisterForm1, step2code string) (string, bool) {
	return Encrypt(TokenModels.SignUpT{
		TokenType:   signUpTokenType,
		MailAddress: form.Mail,
		RememberMe:  form.Remember == "yes",
		Name:        form.Name,
		Password:    form.Password,
		Step2Code:   step2code,
	})
}

func ReadSignUpToken(token string) (*TokenModels.SignUpT, bool) {
	var signUp TokenModels.SignUpT
	tokenDec, ok := StringProcessor.Decrypt(token)
	return &signUp, ok && json.Unmarshal(tokenDec, &signUp) == nil && signUp.TokenType == signUpTokenType
}
