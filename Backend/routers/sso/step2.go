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

func Step2(ctx fiber.Ctx) error {

	encryptedState := ctx.Query(StateQuery)
	encryptedSession := ctx.Cookies(Config.SSOStateInCookie)
	providerName := ctx.Params(ProviderParam)

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+providerName)

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Provider not found: "+providerName)

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
	if session == nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Session type missing for provider: "+providerName)
		return ResponseProcessor.FailurePopup(ctx, SessionInvalid)
	}

	err = StringProcessor.DecryptInterfaceFromB64(encryptedSession, session)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Session decrypt failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SessionDecryptFailed)
	}

	CookieProcessor.DetachSSOCookies(ctx)

	authURL, err := session.GetAuthURL()
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "URL generation failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, AuthURLNotFound)
	}

	parsedAuthURL, err := url.Parse(authURL)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "URL parse failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, URLParseFailed)
	}

	if parsedAuthURL.Query().Get(StateQuery) != encryptedState {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "State mismatch")

		return ResponseProcessor.FailurePopup(ctx, SessionInvalid)
	}

	state, err := TokenProcessor.ReadSSOToken(encryptedState)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SSO token read failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, StateInvalid)
	}
	if state.Provider != providerName {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "SSO provider mismatch: "+state.Provider+" vs "+providerName)
		return ResponseProcessor.FailurePopup(ctx, StateInvalid)
	}

	if RequestProcessor.GetRequestTime(ctx).After(state.Expiry) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "SSO token expired: "+RequestProcessor.GetRequestTime(ctx).Sub(state.Expiry).String())

		return ResponseProcessor.FailurePopup(ctx, SessionExpired)
	}

	values := url.Values{}
	for key, val := range ctx.Queries() {
		values.Add(key, val)
	}

	_, err = session.Authorize(provider, values)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" authorize failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, AuthoriseFailed)
	}

	user, err := provider.FetchUser(session)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" user fetch failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, FetchUserFailed)
	}

	if !StringProcessor.EmailIsValid(user.Email) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), providerName+" sent invalid email: "+user.Email)

		return ResponseProcessor.FailurePopup(ctx, FetchUserFailed)
	}

	exists := true
	blocked := false
	action := "signin"
	userType := UserTypes.All.Viewer.Short
	var userID int32
	var deviceID int16

	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT users.user_id, blocked, type from users where mail = $1 LIMIT 1", user.Email).Scan(&userID, &blocked, &userType)
	if errors.Is(err, pgx.ErrNoRows) {
		exists = false
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Email existence check failed - SQL query: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}

	if !exists {
		action = "signup"
		userID, err = AccountProcessor.RecordNewUser(ctx, userType, "", user.Email, user.Name)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Signup failed: "+err.Error())

			return ResponseProcessor.FailurePopup(ctx, AccountCreateFailed)
		}
		Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed Up: "+strconv.Itoa(int(userID)))
	}

	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account blocked: "+strconv.Itoa(int(userID)))

		return ResponseProcessor.FailurePopup(ctx, AccountBlocked)
	}

	deviceID, err = AccountProcessor.RecordReturningUser(ctx, user.Email, userID, state.Remember, exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Signin failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed In: "+strconv.Itoa(int(userID)))

	token, err := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, state.Remember, state.Provider+"-"+action)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SignInFailed)
	}

	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request completed: "+action+" "+strconv.Itoa(int(userID))+" "+strconv.Itoa(int(deviceID)))
	return ResponseProcessor.SuccessPopup(ctx, token.AccessToken, token.AccessExpires, state.State)
}
