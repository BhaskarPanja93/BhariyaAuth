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

// CreateFreshToken generates a new access + refresh token pair for a user session.
//
// This function is used during initial authentication (signin/signup).
// It creates:
//  1. Access token (short-lived, used for API access).
//  2. Refresh token (long-lived, used to renew access).
//  3. CSRF token (bound to refresh token for protection).
//
// Flow:
//
//	build access token → generate CSRF → build refresh token → return combined
//
// Parameters:
// - userID: authenticated user identifier.
// - deviceID: unique device/session identifier.
// - userType: role or classification of user.
// - remember: whether session should persist longer.
// - identifierType: source of signin (e.g., email, SSO).
//
// Returns:
// - combined struct containing access + refresh tokens and metadata.
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

	// CSRF token bound to refresh token
	combined.CSRF = StringProcessor.SafeString(128)

	combined.RefreshToken, err = StringProcessor.EncryptInterfaceToB64(
		TokenModels.RefreshToken{
			TokenType:      refreshTokenType,
			UserID:         userID,
			DeviceID:       deviceID,
			Visits:         math.MinInt16, // version counter for token rotation
			Created:        now,
			Updated:        now,
			Expiry:         now.Add(Config.RefreshTokenExpireDelta),
			UserType:       userType,
			CSRF:           combined.CSRF,
			Remember:       remember,
			IdentifierType: identifierType,
		},
	)

	return combined, errors.New("Create renew token - refresh: " + err.Error())
}

// CreateRenewToken generates a new access and refresh token from an existing refresh token.
//
// This function is used during token refresh flow.
// It:
//   - Issues a new access token.
//   - Rotates refresh token (new CSRF + incremented version).
//
// Flow:
//
//	validate refresh → create new access → rotate refresh → return updated tokens
//
// Security:
// - Refresh token rotation prevents replay attacks.
// - Visits counter acts as versioning mechanism.
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
	refresh.Visits++ // invalidate previous versions

	combined.CSRF = refresh.CSRF
	combined.RememberMe = refresh.Remember

	combined.RefreshToken, err = StringProcessor.EncryptInterfaceToB64(refresh)

	if err != nil {
		return combined, errors.New("Create renew token - refresh: " + err.Error())
	}
	return combined, nil
}

// CreateMFAToken generates a token for MFA verification flow.
//
// Purpose:
// - Used after OTP validation.
//
// Contains:
// - Step2Code (OTP reference)
// - User + Device identity
// - Creation timestamp
// - Verified flag
func CreateMFAToken(ctx fiber.Ctx, userID int32, deviceID int16, step2code string) (string, error) {

	return StringProcessor.EncryptInterfaceToB64(
		TokenModels.MFAToken{
			TokenType: mfaTokenType,
			Step2Code: step2code,
			UserID:    userID,
			DeviceID:  deviceID,
			Created:   RequestProcessor.GetRequestTime(ctx),
			Verified:  true,
		},
	)
}

// CreateSSOToken generates state token for SSO flow.
//
// Purpose:
// - Maintains state across OAuth redirects.
// - Encodes provider + expiry + remember flag.
//
// Used in:
// - SSO initiation → callback validation.
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

// CreatePasswordResetToken generates token for password reset flow.
//
// Contains:
// - User identity
// - Email
// - OTP reference
//
// Used in:
// - Step1 → Step2 password reset verification.
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

// CreateSignInToken generates token for signin flow.
//
// Contains:
// - User identity
// - Selected authentication method (OTP/password)
// - OTP reference (if applicable)
// - Remember flag
//
// Used in:
// - SignIn Step1 → Step2
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

// CreateSignUpToken generates token for multi-step registration flow.
//
// Contains:
// - User input data (name, email, password)
// - OTP reference
// - Remember flag
//
// Used in:
// - Signup Step1 → Step2
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
