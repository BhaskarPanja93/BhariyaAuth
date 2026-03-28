package sso

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	ResponseProcessor "BhariyaAuth/processors/response"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"database/sql"
	"errors"
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
	// Initialize all providers usable for SSO
	goth.UseProviders(
		google.New(
			Secrets.GoogleClientId,
			Secrets.GoogleClientSecret,
			Config.ServerSSOCallbackURL+"/google",
			"profile", "email"),
		discord.New(
			Secrets.DiscordClientId,
			Secrets.DiscordClientSecret,
			Config.ServerSSOCallbackURL+"/discord",
			"identify", "email", "openid"),
		microsoftonline.New(
			Secrets.MicrosoftClientId,
			Secrets.MicrosoftClientSecret,
			Config.ServerSSOCallbackURL+"/microsoftonline",
			"user.read"),
	)
}

// Step1 takes in the required SSO provider as URL segment `/sso/google`.
// It then creates a struct with provider name, a supplied expiry and if the session should be remembered.
// This prepared struct (after marshal and encryption) will serve as state for the entire SSO flow.
// The state is packed into goth session which is then serialized and attached nto the request as SSO session cookie
// before redirecting the request to the goth provided URL
func Step1(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	// Processor names must match the ones in goth
	providerName := ctx.Params(ProviderParam)
	// Fetch provider from goth
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		// Provider not found
		return ResponseProcessor.SSOFailurePopup(ctx, UnknownProvider)
	}
	// State struct for the SSO flow
	state := TokenModels.SSOStateT{
		Provider:   providerName,
		Expiry:     now.Add(Config.SSOCookieExpireDelta),
		RememberMe: ctx.Query("remember", "no") == "yes",
	}
	stateMarshal, err := json.Marshal(state)
	if err != nil {
		// State Marshaling failed
		return ResponseProcessor.SSOFailurePopup(ctx, StateMarshalFailed)
	}
	encryptedState, ok := StringProcessor.Encrypt(stateMarshal)
	if !ok {
		// State Marshal encrypting failed
		return ResponseProcessor.SSOFailurePopup(ctx, StateEncryptFailed)
	}
	// Create SSO session in goth that will be attached to cookie
	session, err := provider.BeginAuth(encryptedState)
	if err != nil {
		// Session creation failed
		return ResponseProcessor.SSOFailurePopup(ctx, BeginFailed)
	}
	// Fetch URL to forward to
	sendURL, err := session.GetAuthURL()
	if err != nil {
		// URL fetch failed
		return ResponseProcessor.SSOFailurePopup(ctx, AuthURLNotFound)
	}
	sessionMarshal, err := json.Marshal(session)
	if err != nil {
		// Session marshaling failed
		return ResponseProcessor.SSOFailurePopup(ctx, SessionMarshalFailed)
	}
	enc, ok := StringProcessor.Encrypt(sessionMarshal)
	if !ok {
		// Session marshal encrypting failed
		return ResponseProcessor.SSOFailurePopup(ctx, SessionEncryptFailed)
	}
	CookieProcessor.AttachSSOCookie(ctx, enc)
	// Redirect the request to the provider's URL with state as param and session as cookie attached
	return ctx.Redirect().To(sendURL)
}

// Step2 receives the state and provider name as provided by the auth provider as well as the cookie present in the request.
// The session from cookie is first decrypted and the embedded state is matched against the state received in URL.
// Only then the URL string is unmarshalled, checked for staleness and used to authorize and fetch the user details using goth.
// After a valid email is received, the same is checked in self databases for existing user.
// If no user matches the email, a new one is created, else old account is used to log in.
// On success, the auth token, its expiry is passed into the embedded HTML that `should` have acted as a popup,
// passing on the values using JavaScript postMessage
func Step2(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	// Fetch provider name, state and session values from URL and cookie
	stateString := ctx.Query("state")
	sessionString := ctx.Cookies(Config.SSOStateInCookie)
	providerName := ctx.Params(ProviderParam)
	// Fetch the provider from the URL
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		// Provider fetch failed
		return ResponseProcessor.SSOFailurePopup(ctx, UnknownProvider)
	}
	CookieProcessor.DetachSSOCookies(ctx)
	// Decrypt received session
	sessDec, ok := StringProcessor.Decrypt(sessionString)
	if !ok {
		// Session Decrypt failed
		return ResponseProcessor.SSOFailurePopup(ctx, SessionDecryptFailed)
	}
	// Unmarshal received session
	sess, err := provider.UnmarshalSession(string(sessDec))
	if err != nil {
		// Session Unmarshalling failed
		return ResponseProcessor.SSOFailurePopup(ctx, SessionUnmarshalFailed)
	}
	// Fetch the Redirect URL from session cookie
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		// URL fetch failed
		return ResponseProcessor.SSOFailurePopup(ctx, AuthURLNotFound)
	}
	// Fetch state according to the URL
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		// URL parse failed
		return ResponseProcessor.SSOFailurePopup(ctx, URLParseFailed)
	}
	// State embedded in session must match the state received as URL param
	originalState := authURL.Query().Get("state")
	if originalState != stateString {
		return ResponseProcessor.SSOFailurePopup(ctx, SessionInvalid)
	}
	// Decrypt received state
	stateDec, ok := StringProcessor.Decrypt(stateString)
	if !ok {
		// State Decrypt failed
		return ResponseProcessor.SSOFailurePopup(ctx, StateDecryptFailed)
	}
	state := TokenModels.SSOStateT{}
	err = json.Unmarshal(stateDec, &state)
	if err != nil {
		// State Unmarshalling failed
		return ResponseProcessor.SSOFailurePopup(ctx, StateUnMarshalFailed)
	}
	// Check if the state is fresh
	if now.After(state.Expiry) {
		return ResponseProcessor.SSOFailurePopup(ctx, SessionExpired)
	}
	// Fetch all URL params received from the provider
	values := url.Values{}
	for key, val := range ctx.Queries() {
		values.Add(key, val)
	}
	// Pass all the URL params into goth for authorizing
	_, err = sess.Authorize(provider, values)
	if err != nil {
		// Authorize failed
		return ResponseProcessor.SSOFailurePopup(ctx, AuthoriseFailed)
	}
	// Fetch session using the session
	user, err := provider.FetchUser(sess)
	if err != nil {
		// Fetch user failed
		return ResponseProcessor.SSOFailurePopup(ctx, FetchUserFailed)
	}
	if !StringProcessor.EmailIsValid(user.Email) {
		// Email invalid
		return ResponseProcessor.SSOFailurePopup(ctx, FetchUserFailed)
	}
	exists := true
	blocked := false
	action := "login"
	userType := UserTypes.All.Viewer.Short
	var userID int32
	var deviceID int16
	// Prefetch all data in a single query to prevent DB overloading with multiple requests
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT users.user_id, blocked, type from users where mail = $1", user.Email).Scan(&userID, &blocked, &userType)
	if errors.Is(err, sql.ErrNoRows) { // Account doesn't exist
		exists = false
	} else if err != nil { // Other DB error
		return ResponseProcessor.SSOFailurePopup(ctx, LoginFailed)
	}
	// Create new account, same as register handler
	if !exists {
		action = "register"
		userID, ok = AccountProcessor.RecordNewUser(ctx, userType, "", user.Email, user.Name)
		if !ok {
			// Registration failed
			return ResponseProcessor.SSOFailurePopup(ctx, AccountCreateFailed)
		}
	}
	if blocked {
		// Account found and is blocked from logging in
		return ResponseProcessor.SSOFailurePopup(ctx, AccountBlocked)
	}
	// Login new device regardless of new or old account
	deviceID, ok = AccountProcessor.RecordReturningUser(ctx, user.Email, userID, state.RememberMe, exists)
	if !ok {
		// login failed
		return ResponseProcessor.SSOFailurePopup(ctx, LoginFailed)
	}
	// Create access and refresh tokens
	token, ok := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, state.RememberMe, state.Provider+"-"+action)
	if !ok {
		// Token creation failed
		return ResponseProcessor.SSOFailurePopup(ctx, LoginFailed)
	}
	// Attach the refresh token as cookie
	CookieProcessor.AttachAuthCookies(ctx, token)
	// Return access token and its expiry as embedded text in HTML
	return ResponseProcessor.SSOSuccessPopup(ctx, token.AccessToken, token.AccessExpires)
}
