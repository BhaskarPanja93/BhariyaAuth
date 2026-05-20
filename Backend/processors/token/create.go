package token

import (
	Config "BhariyaAuth/constants/config"
	FormModels "BhariyaAuth/models/requests"
	TokenModels "BhariyaAuth/models/tokens"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	"errors"
	"math"

	"github.com/gofiber/fiber/v3"
)

func CreateFreshToken(
	ctx fiber.Ctx,
	userID int32,
	deviceID int16,
	userType string,
	remember bool,
	identifierType string,
) (TokenModels.NewTokenCombined, error) {

	now := RequestProcessor.GetRequestTime(ctx)

	var combined TokenModels.NewTokenCombined
	var err error

	combined.AccessExpires = now.Add(Config.AccessTokenExpireDelta)

	combined.AccessToken, err = StringProcessor.EncryptInterfaceToB64(
		TokenModels.AccessToken{
			TokenType: accessTokenType,
			UserID:    userID,
			DeviceID:  deviceID,
			UserType:  userType,
			Expiry:    combined.AccessExpires,
			Remember:  remember,
		},
	)
	if err != nil {
		return combined, errors.New("Create renew token - access: " + err.Error())
	}

	combined.RememberMe = remember

	combined.CSRF = StringProcessor.SafeString(128)

	combined.RefreshToken, err = StringProcessor.EncryptInterfaceToB64(
		TokenModels.RefreshToken{
			TokenType:      refreshTokenType,
			UserID:         userID,
			DeviceID:       deviceID,
			Visits:         math.MinInt16,
			Created:        now,
			Updated:        now,
			Expiry:         now.Add(Config.RefreshTokenExpireDelta),
			UserType:       userType,
			CSRF:           combined.CSRF,
			Remember:       remember,
			IdentifierType: identifierType,
		},
	)
	if err != nil {
		return combined, errors.New("Create renew token - refresh: " + err.Error())
	}

	return combined, nil
}

func CreateRenewToken(
	ctx fiber.Ctx,
	refresh TokenModels.RefreshToken,
) (TokenModels.NewTokenCombined, error) {

	now := RequestProcessor.GetRequestTime(ctx)

	var combined TokenModels.NewTokenCombined
	var err error

	combined.AccessToken, err = StringProcessor.EncryptInterfaceToB64(
		TokenModels.AccessToken{
			TokenType: accessTokenType,
			UserID:    refresh.UserID,
			DeviceID:  refresh.DeviceID,
			UserType:  refresh.UserType,
			Expiry:    now.Add(Config.AccessTokenExpireDelta),
			Remember:  refresh.Remember,
		},
	)
	if err != nil {
		return combined, errors.New("Create renew token - access: " + err.Error())
	}

	refresh.CSRF = StringProcessor.SafeString(128)
	refresh.Expiry = now.Add(Config.RefreshTokenExpireDelta)
	refresh.Visits++

	combined.CSRF = refresh.CSRF
	combined.RememberMe = refresh.Remember

	combined.RefreshToken, err = StringProcessor.EncryptInterfaceToB64(refresh)

	if err != nil {
		return combined, errors.New("Create renew token - refresh: " + err.Error())
	}
	return combined, nil
}

func CreateMFAToken(ctx fiber.Ctx, userID int32, deviceID int16, step2code string, verified bool) (string, error) {
	code := step2code
	if verified {
		code = ""
	}

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.MFAToken{
			TokenType: mfaTokenType,
			Step2Code: code,
			UserID:    userID,
			DeviceID:  deviceID,
			Created:   RequestProcessor.GetRequestTime(ctx),
			Verified:  verified,
		},
	)
}

func CreateSSOToken(ctx fiber.Ctx, provider string, state string, remember bool) (string, error) {

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.SSOState{
			TokenType: ssoTokenType,
			Provider:  provider,
			State:     state,
			Expiry:    RequestProcessor.GetRequestTime(ctx).Add(Config.SSOCookieExpireDelta),
			Remember:  remember,
		},
	)
}

func CreatePasswordResetToken(mail string, userID int32, step2code string) (string, error) {

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.PasswordReset{
			TokenType:   passwordResetTokenType,
			MailAddress: mail,
			UserID:      userID,
			Step2Code:   step2code,
		},
	)
}

func CreateSignInToken(
	form *FormModels.SignInForm1,
	userID int32,
	step2process string,
	step2code string,
) (string, error) {

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.SignIn{
			TokenType:    signInTokenType,
			UserID:       userID,
			Remember:     form.Remember == "yes",
			Step2Process: step2process,
			Step2Code:    step2code,
			MailAddress:  form.Mail,
		},
	)
}

func CreateSignUpToken(form *FormModels.SignUpForm1, step2code string) (string, error) {

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.SignUp{
			TokenType:   signUpTokenType,
			MailAddress: form.Mail,
			Remember:    form.Remember == "yes",
			Name:        form.Name,
			Password:    form.Password,
			Step2Code:   step2code,
		},
	)
}
