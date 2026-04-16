package sso

import (
	Config "BhariyaAuth/constants/config"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	ResponseProcessor "BhariyaAuth/processors/sso"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
)

const step2FileName = "routers/sso/step2"

// Step2 handles the callback phase of the Single Sign-On (SSO) authentication flow.
//
// This function is invoked after the external SSO provider redirects back to the application.
// It validates the integrity of the SSO flow, retrieves user information, and completes signin/signup.
//
// Core Responsibilities:
//  1. Extract and validate the provider, state, and session from request.
//  2. Decrypt and restore the session stored in cookies.
//  3. Verify that the state returned by the provider matches the original state (CSRF protection).
//  4. Decrypt and validate the state payload (expiry, type).
//  5. Complete authorization with the provider using returned query parameters.
//  6. Fetch authenticated user details from the provider.
//  7. Validate user identity (email).
//  8. Check if user exists in the database:
//     - If not, signup as a new user.
//     - If yes, proceed with signin.
//  9. Handle blocked users.
//  10. Record device/session and generate authentication tokens.
//  11. Return tokens to frontend via secure popup communication.
//
// Security Considerations:
// - State validation prevents CSRF attacks.
// - Encrypted session + state ensures tamper resistance.
// - Expiry check prevents replay attacks.
// - Cookie is cleared early to prevent reuse.
// - Email validation ensures identity integrity.
//
// Returns:
// - HTML popup response with access token on success.
// - Failure popup response for any validation or processing error.
func Step2(ctx fiber.Ctx) error {

	// Extract inputs from request
	encryptedState := ctx.Query(StateQuery)                  // Encrypted state returned by provider
	encryptedSession := ctx.Cookies(Config.SSOStateInCookie) // Encrypted session stored earlier
	providerName := ctx.Params(ProviderParam)                // Provider identifier

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested: "+providerName)

	// Retrieve provider configuration from Goth
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Provider not found: "+providerName)

		// Invalid or unsupported provider
		return ResponseProcessor.FailurePopup(ctx, UnknownProvider)
	}

	var session goth.Session
	switch providerName {
	case googleProvider.Name():
		session = &google.Session{}
	case discordProvider.Name():
		session = &discord.Session{}
	case microsoftonlineProvider.Name():
		session = &microsoftonline.Session{}
	default:
		session = nil
	}

	err = StringProcessor.DecryptInterfaceFromB64(encryptedSession, session)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Session decrypt failed: "+err.Error())

		// Session decrypt failed
		return ResponseProcessor.FailurePopup(ctx, SessionDecryptFailed)
	}

	// Immediately clear SSO cookies
	CookieProcessor.DetachSSOCookies(ctx)

	// Retrieve original authentication URL (contains original state)
	authURL, err := session.GetAuthURL()
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "URL generation failed: "+err.Error())

		// URL fetch failed
		return ResponseProcessor.FailurePopup(ctx, AuthURLNotFound)
	}

	// Parse URL to extract embedded state
	parsedAuthURL, err := url.Parse(authURL)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "URL parse failed: "+err.Error())

		// URL parse failed
		return ResponseProcessor.FailurePopup(ctx, URLParseFailed)
	}

	// Validate returned state matches original (critical CSRF protection)
	if parsedAuthURL.Query().Get(StateQuery) != encryptedState {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "State mismatch")

		return ResponseProcessor.FailurePopup(ctx, SessionInvalid)
	}

	// Decrypt state payload received from provider
	state, err := TokenProcessor.ReadSSOToken(ctx.Query("state"))
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SSO token read failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, StateInvalid)
	}

	// Validate state freshness (prevents replay attacks)
	if RequestProcessor.GetRequestTime(ctx).After(state.Expiry) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "SSO token expired: "+RequestProcessor.GetRequestTime(ctx).Sub(state.Expiry).String())

		return ResponseProcessor.FailurePopup(ctx, SessionExpired)
	}

	// Collect all query parameters returned by provider
	// These include auth codes, tokens, and other provider-specific data
	values := url.Values{}
	for key, val := range ctx.Queries() {
		values.Add(key, val)
	}

	// Complete authorization flow using provider SDK
	_, err = session.Authorize(provider, values)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" authorize failed: "+err.Error())

		// Authorize failed
		return ResponseProcessor.FailurePopup(ctx, AuthoriseFailed)
	}

	// Fetch authenticated user profile from provider
	user, err := provider.FetchUser(session)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" user fetch failed: "+err.Error())

		// Fetch user failed
		return ResponseProcessor.FailurePopup(ctx, FetchUserFailed)
	}

	// Validate email (critical identity field)
	if !StringProcessor.EmailIsValid(user.Email) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" sent invalid email: "+user.Email)

		// Email invalid
		return ResponseProcessor.FailurePopup(ctx, FetchUserFailed)
	}

	// Initialize user state variables
	exists := true
	blocked := false
	action := "signin"
	userType := UserTypes.All.Viewer.Short
	var userID int32
	var deviceID int16

	// Attempt to fetch user record from database
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT users.user_id, blocked, type from users where mail = $1 LIMIT 1", user.Email).Scan(&userID, &blocked, &userType)
	if errors.Is(err, pgx.ErrNoRows) {
		// No existing account found → mark for registration
		exists = false
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Email existence check failed - SQL query: "+err.Error())

		// Unexpected DB error
		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}

	// Handle new user registration
	if !exists {
		action = "signup"
		userID, err = AccountProcessor.RecordNewUser(ctx, userType, "", user.Email, user.Name)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Signup failed: "+err.Error())

			// Registration failed
			return ResponseProcessor.FailurePopup(ctx, AccountCreateFailed)
		}
		Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed Up: "+user.Email+" "+strconv.Itoa(int(userID)))
	}

	// Prevent signin if account is blocked
	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account blocked: "+strconv.Itoa(int(userID)))

		return ResponseProcessor.FailurePopup(ctx, AccountBlocked)
	}

	// Record device/session for both new and returning users
	deviceID, err = AccountProcessor.RecordReturningUser(ctx, user.Email, userID, state.Remember, exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Signin failed: "+err.Error())

		// signin failed
		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed In: "+strconv.Itoa(int(userID)))

	// Generate authentication tokens (access + refresh)
	token, err := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, state.Remember, state.Provider+"-"+action)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		// Token creation failed
		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}

	// Attach refresh token securely via cookies
	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request completed: "+action+" "+strconv.Itoa(int(userID))+" "+strconv.Itoa(int(deviceID)))
	// Return access token via popup response (used by frontend)
	return ResponseProcessor.SuccessPopup(ctx, token.AccessToken, token.AccessExpires, state.State)
}
