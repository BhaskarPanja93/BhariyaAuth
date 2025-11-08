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
	stateString := ctx.Query("state")
	sessionString := ctx.Cookies(Config.SSOStateInCookie)
	ResponseProcessor.DetachSSOCookies(ctx)
	if sessionString == "" || stateString == "" {
		return ResponseProcessor.SSOFailureResponse(ctx, "sessionString/stateString empty")
	}
	stateDec, ok1 := StringProcessor.Decrypt(stateString)
	sessDec, ok2 := StringProcessor.Decrypt(sessionString)
	if !ok1 || !ok2 {
		return ResponseProcessor.SSOFailureResponse(ctx, "Unable to decrypt session and state")
	}
	state := TokenModels.SSOStateT{}
	err := json.Unmarshal(stateDec, &state)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 State Unmarshal Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Unmarshal state failed. Please try again")
	}
	provider, err := goth.GetProvider(state.Provider)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 GetProvider Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Provider unknown")
	}
	sess, err := provider.UnmarshalSession(string(sessDec))
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 UnmarshalSession Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Unmarshal session failed. Please try again")
	}
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 GetAuthURL Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Get Auth URL failed")
	}
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 URLParse Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "URL parse failed")
	}
	originalState := authURL.Query().Get("state")
	if originalState != stateString {
		Logger.IntentionalFailure("SSO-2 State doesnt match session")
		return ResponseProcessor.SSOFailureResponse(ctx, "State does not match")
	}
	if state.Expiry.Before(time.Now()) {
		Logger.IntentionalFailure("SSO-2 State too old")
		return ResponseProcessor.SSOFailureResponse(ctx, "State expired")
	}
	values := url.Values{}
	for key, val := range ctx.Queries() {
		values.Add(key, val)
	}
	_, err = sess.Authorize(provider, values)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 Authorize Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Authorize failed")
	}
	user, err := provider.FetchUser(sess)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-2 FetchUser Failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Fetch user failed")
	}
	userID, found := AccountProcessor.GetIDFromMail(user.Email)
	if !found {
		userID = Generators.UserID()
		refreshID := Generators.RefreshID()
		if !AccountProcessor.RecordNewUser(userID, "", user.Email, user.Name) {
			Logger.AccidentalFailure(fmt.Sprintf("SSO-2 RecordNew Failed [%d-%s]", userID, user.Email))
			return ResponseProcessor.SSOFailureResponse(ctx, "Failed to create account, please contact support")
		}
		if !AccountProcessor.RecordReturningUser(refreshID, userID, state.RememberMe) {
			Logger.AccidentalFailure(fmt.Sprintf("SSO-2 RecordNewReturning Failed [%d-%s]", userID, user.Email))
			return ResponseProcessor.SSOFailureResponse(ctx, "Failed to login, please try again or contact support")
		}
		token := TokenProcessor.CreateFreshToken(
			userID,
			refreshID,
			UserTypes.All.Viewer,
			state.RememberMe,
			fmt.Sprintf("%s-register", state.Provider),
		)
		ResponseProcessor.AttachAuthCookies(ctx, token)
		Logger.Success(fmt.Sprintf("SSO-2 Registered: [%d-%d-%s]", userID, refreshID, user.Email))
		return ResponseProcessor.SSOSuccessResponse(ctx, token.AccessToken, state.FrontendState, state.Origin)
	} else {
		if AccountProcessor.UserIsBlacklisted(userID) {
			Logger.IntentionalFailure(fmt.Sprintf("SSO-2 Blacklisted account [%d-%s] attempted login", userID, user.Email))
			return ResponseProcessor.SSOFailureResponse(ctx, "Your account is disabled, please contact support")
		}
		refreshID := Generators.RefreshID()
		if !AccountProcessor.RecordReturningUser(refreshID, userID, state.RememberMe) {
			Logger.AccidentalFailure(fmt.Sprintf("SSO-2 RecordReturning Failed [%d-%s]", userID, user.Email))
			return ResponseProcessor.SSOFailureResponse(ctx, "Failed to login, please try again or contact support")
		}
		token := TokenProcessor.CreateFreshToken(
			userID,
			refreshID,
			UserTypes.Find(AccountProcessor.GetUserType(userID)),
			state.RememberMe,
			fmt.Sprintf("%s-login", state.Provider),
		)
		ResponseProcessor.AttachAuthCookies(ctx, token)
		Logger.Success(fmt.Sprintf("SSO-2 LoggedIn: [%d-%d-%s]", userID, refreshID, user.Email))
		return ResponseProcessor.SSOSuccessResponse(ctx, token.AccessToken, state.FrontendState, state.Origin)
	}
}

func Step1(ctx fiber.Ctx) error {
	processor := ctx.Params("processor")
	provider, err := goth.GetProvider(processor)
	if err != nil {
		return ResponseProcessor.SSOFailureResponse(ctx, "Provider unknown")
	}
	state := TokenModels.SSOStateT{
		Provider:      processor,
		Expiry:        time.Now().Add(Config.SSOCookieExpireDelta),
		FrontendState: ctx.Query("state", ""),
		Origin:        ctx.Query("origin", ""),
		RememberMe:    ctx.Query("remember_me", "no") == "yes",
	}
	stateMarshal, err := json.Marshal(state)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-1 State Marshal failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Marshal state failed. Please try again")
	}
	encryptedState, ok := StringProcessor.Encrypt(stateMarshal)
	if !ok {
		Logger.AccidentalFailure("SSO-1 State Encrypt failed")
		return ResponseProcessor.SSOFailureResponse(ctx, "Encrypt state failed. Please try again")
	}
	sess, err := provider.BeginAuth(encryptedState)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-1 BeginAuth failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Begin auth failed. Please try again")
	}
	sendURL, err := sess.GetAuthURL()
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-1 GetAuthURL failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Get Auth URL failed. Please try again")
	}
	marshal, err := json.Marshal(sess)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("SSO-1 Sess Marshal failed: %s", err.Error()))
		return ResponseProcessor.SSOFailureResponse(ctx, "Marshal session failed. Please try again")
	}
	enc, ok := StringProcessor.Encrypt(marshal)
	if !ok {
		Logger.AccidentalFailure("SSO-1 Sess Encrypt failed")
		return ResponseProcessor.SSOFailureResponse(ctx, "Encrypt session failed. Please try again")
	}
	ResponseProcessor.AttachSSOCookie(ctx, enc)
	return ctx.Redirect().To(sendURL)
}
