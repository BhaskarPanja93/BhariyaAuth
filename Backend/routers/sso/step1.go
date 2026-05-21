package sso

import (
	CookieProcessor "BhariyaAuth/processors/cookies"
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	ResponseProcessor "BhariyaAuth/processors/sso"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"

	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
)

const step1FileName = "routers/sso/step1"

func Step1(ctx fiber.Ctx) error {

	providerName := ctx.Params(ProviderParam)
	state := ctx.Query(StateQuery)
	remember := ctx.Query("remember", "no") == "yes"
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+providerName)

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Provider not found: "+providerName)

		return ResponseProcessor.FailurePopup(ctx, UnknownProvider)
	}

	encryptedState, err := TokenProcessor.CreateSSOToken(ctx, providerName, state, remember)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "State token creation failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, StateEncryptFailed)
	}

	session, err := provider.BeginAuth(encryptedState)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Session creation failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SessionCreateFailed)
	}

	authURL, err := session.GetAuthURL()
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "URL generation failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, AuthURLNotFound)
	}

	encryptedSession, err := StringProcessor.EncryptInterfaceToB64(session)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Session encrypt failed: "+err.Error())

		return ResponseProcessor.FailurePopup(ctx, SessionEncryptFailed)
	}

	CookieProcessor.AttachSSOCookie(ctx, encryptedSession)

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request Complete")
	return ctx.Redirect().To(authURL)
}
