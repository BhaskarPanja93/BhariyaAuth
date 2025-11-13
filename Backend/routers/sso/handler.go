package sso

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	ResponseProcessor "BhariyaAuth/processors/response"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	"fmt"

	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
)

func init() {
	goth.UseProviders(
		google.New(
			Secrets.GoogleClientId,
			Secrets.GoogleClientSecret,
			Config.ServerSSOCallbackURL,
			"profile", "email"),
		discord.New(
			Secrets.DiscordClientId,
			Secrets.DiscordClientSecret,
			Config.ServerSSOCallbackURL,
			"identify", "email", "openid"),
		microsoftonline.New(
			Secrets.MicrosoftClientId,
			Secrets.MicrosoftClientSecret,
			Config.ServerSSOCallbackURL,
			"user.read"),
	)
}

func Step2(ctx fiber.Ctx) error {
	now := time.Now().UTC()
	stateString := ctx.Query("state")
	sessionString := ctx.Cookies(Config.SSOStateInCookie)
	ResponseProcessor.DetachSSOCookies(ctx)
	if sessionString == "" || stateString == "" {
		return ResponseProcessor.SSOFailurePopup(ctx, "Session not found")
	}
	stateDec, ok1 := StringProcessor.Decrypt(stateString)
	sessDec, ok2 := StringProcessor.Decrypt(sessionString)
	if !ok1 || !ok2 {
		return ResponseProcessor.SSOFailurePopup(ctx, "Decrypt issue. Please try again")
	}
	state := TokenModels.SSOStateT{}
	err := json.Unmarshal(stateDec, &state)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] State Unmarshal error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "State Parser issue. Please try again")
	}
	provider, err := goth.GetProvider(state.Provider)
	if err != nil {
		return ResponseProcessor.SSOFailurePopup(ctx, "Provider unknown")
	}
	sess, err := provider.UnmarshalSession(string(sessDec))
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] Session Unmarshal error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Session Parser issue. Please try again")
	}
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] GetAuthURL error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Get Auth URL issue. Please try again")
	}
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] URLParse error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "URL parse failed. Please try again")
	}
	originalState := authURL.Query().Get("state")
	if originalState != stateString {
		return ResponseProcessor.SSOFailurePopup(ctx, "Invalid session")
	}
	if now.After(state.Expiry) {
		return ResponseProcessor.SSOFailurePopup(ctx, "Session expired")
	}
	values := url.Values{}
	for key, val := range ctx.Queries() {
		values.Add(key, val)
	}
	_, err = sess.Authorize(provider, values)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] Authorize error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Authorize failed")
	}
	user, err := provider.FetchUser(sess)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO2] FetchUser error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Fetch user failed")
	}
	userID, found := AccountProcessor.GetIDFromMail(user.Email)
	if !found {
		userID = Generators.UserID()
		refreshID := Generators.RefreshID()
		if !AccountProcessor.RecordNewUser(userID, "", user.Email, user.Name) {
			Logger.AccidentalFailure(fmt.Sprintf("[SSO2] RecordNew Failed for [MAIL-%s]", user.Email))
			return ResponseProcessor.SSOFailurePopup(ctx, "Failed to create account, please contact support")
		}
		if !AccountProcessor.RecordReturningUser(user.Email, ctx.Get("User-Agent"), refreshID, userID, state.RememberMe) {
			Logger.AccidentalFailure(fmt.Sprintf("[SSO2] RecordNewReturning Failed for [UID-%d]", userID))
			return ResponseProcessor.SSOFailurePopup(ctx, "Failed to login, please try again or contact support")
		}
		token, ok := TokenProcessor.CreateFreshToken(
			userID,
			refreshID,
			UserTypes.All.Viewer,
			state.RememberMe,
			fmt.Sprintf("%s-register", state.Provider),
		)
		if !ok {
			return ResponseProcessor.SSOFailurePopup(ctx, "Could not create token. Please try again")
		}
		ResponseProcessor.AttachAuthCookies(ctx, token)
		Logger.Success(fmt.Sprintf("[SSO2] Registered: [UID-%d-RID-%d-MAIL-%s]", userID, refreshID, user.Email))
		return ResponseProcessor.SSOSuccessPopup(ctx, token.AccessToken, state.Origin)
	} else {
		if AccountProcessor.CheckUserIsBlacklisted(userID) {
			Logger.IntentionalFailure(fmt.Sprintf("[SSO2] Blacklisted account [UID-%d]", userID))
			return ResponseProcessor.SSOFailurePopup(ctx, "Your account is disabled, please contact support")
		}
		refreshID := Generators.RefreshID()
		if !AccountProcessor.RecordReturningUser(user.Email, ctx.Get("User-Agent"), refreshID, userID, state.RememberMe) {
			Logger.AccidentalFailure(fmt.Sprintf("[SSO2] RecordNewReturning Failed for [UID-%d]", userID))
			return ResponseProcessor.SSOFailurePopup(ctx, "Failed to login, please try again or contact support")
		}
		token, ok := TokenProcessor.CreateFreshToken(
			userID,
			refreshID,
			AccountProcessor.GetUserType(userID),
			state.RememberMe,
			fmt.Sprintf("%s-login", state.Provider),
		)
		if !ok {
			return ResponseProcessor.SSOFailurePopup(ctx, "Could not create token. Please try again")
		}
		ResponseProcessor.AttachAuthCookies(ctx, token)
		Logger.Success(fmt.Sprintf("[SSO2] LoggedIn: [UID-%d-RID-%d-MAIL-%s]", userID, refreshID, user.Email))
		return ResponseProcessor.SSOSuccessPopup(ctx, token.AccessToken, state.Origin)
	}
}

func Step1(ctx fiber.Ctx) error {
	now := time.Now().UTC()
	processor := ctx.Params("processor")
	provider, err := goth.GetProvider(processor)
	if err != nil {
		return ResponseProcessor.SSOFailurePopup(ctx, "Provider unknown")
	}
	state := TokenModels.SSOStateT{
		Provider:   processor,
		Expiry:     now.Add(Config.SSOCookieExpireDelta),
		Origin:     ctx.Query("origin", ""),
		RememberMe: ctx.Query("remember_me", "no") == "yes",
	}
	stateMarshal, err := json.Marshal(state)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO1] State Marshal failed reason: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "State Parse issue. Please try again")
	}
	encryptedState, ok := StringProcessor.Encrypt(stateMarshal)
	if !ok {
		Logger.AccidentalFailure("[SSO1] State Encrypt failed")
		return ResponseProcessor.SSOFailurePopup(ctx, "State Encrypt issue. Please try again")
	}
	sess, err := provider.BeginAuth(encryptedState)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO1] BeginAuth error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Begin auth issue. Please try again")
	}
	sendURL, err := sess.GetAuthURL()
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO1] GetAuthURL error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Get Auth URL issue. Please try again")
	}
	marshal, err := json.Marshal(sess)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[SSO1] Session Marshal error: %s", err.Error()))
		return ResponseProcessor.SSOFailurePopup(ctx, "Session Parse issue. Please try again")
	}
	enc, ok := StringProcessor.Encrypt(marshal)
	if !ok {
		Logger.AccidentalFailure("[SSO1] Session Encrypt failed")
		return ResponseProcessor.SSOFailurePopup(ctx, "Session Encrypt issue. Please try again")
	}
	ResponseProcessor.AttachSSOCookie(ctx, enc)
	return ctx.Redirect().To(sendURL)
}
