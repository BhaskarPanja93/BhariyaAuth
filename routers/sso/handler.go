package sso

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserModels "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	"BhariyaAuth/processors/generator"
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

func Step1(ctx fiber.Ctx) error {
	processor := ctx.Params("processor")
	provider, err := goth.GetProvider(processor)
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.InvalidEntries,
			ResponseModels.DefaultAuth,
			[]string{"Provider Unknown"},
			nil,
			nil,
		))
	}

	state := TokenModels.SSOStateT{
		Provider:      processor,
		Expiry:        time.Now().Add(time.Minute * 10),
		FrontendState: ctx.Query("state", ""),
		Origin:        ctx.Query("origin", ""),
		RememberMe:    ctx.Query("remember_me", "no") == "yes",
	}
	stateMarshal, err := json.Marshal(state)
	if err == nil {
		encryptedState, ok := StringProcessor.Encrypt(stateMarshal)
		if ok {
			sess, err := provider.BeginAuth(encryptedState)
			if err == nil {
				sendURL, err := sess.GetAuthURL()
				if err == nil {
					marshal, err := json.Marshal(sess)
					if err == nil {
						enc, ok := StringProcessor.Encrypt(marshal)
						if ok {
							ctx.Cookie(&fiber.Cookie{
								Name:     Config.SSOStateInCookie,
								Value:    enc,
								Expires:  time.Now().Add(time.Minute * 15),
								Domain:   Config.CookieDomain,
								SameSite: fiber.CookieSameSiteStrictMode,
								Secure:   true,
								HTTPOnly: true,
							})
							return ctx.Redirect().To(sendURL)
						}
					}
				}
			}
		}
	}
	return ResponseProcessor.SSOFailureResponse(ctx, "This didn't work, maybe try again??")
}

func Step2(ctx fiber.Ctx) error {
	stateString := ctx.Query("state")
	sessionString := ctx.Cookies(Config.SSOStateInCookie)
	if stateString == "" || sessionString == "" {
		stateDec, ok1 := StringProcessor.Decrypt(stateString)
		sessDec, ok2 := StringProcessor.Decrypt(sessionString)
		if ok1 && ok2 {
			state := TokenModels.SSOStateT{}
			err := json.Unmarshal(stateDec, &state)
			if err == nil {
				provider, err := goth.GetProvider(state.Provider)
				if err == nil {
					sess, err := provider.UnmarshalSession(string(sessDec))
					if err == nil {
						rawAuthURL, err := sess.GetAuthURL()
						if err == nil {
							authURL, err := url.Parse(rawAuthURL)
							if err == nil {
								originalState := authURL.Query().Get("state")
								if originalState == stateString && state.Expiry.After(time.Now()) {
									values := url.Values{}
									ctx.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
										values.Add(string(key), string(value))
									})
									_, err = sess.Authorize(provider, values)
									if err == nil {
										user, err := provider.FetchUser(sess)
										if err == nil {
											userID, found := AccountProcessor.GetIDFromMail(user.Email)
											if found {
												if AccountProcessor.UserIsBlacklisted(userID) {
													return ResponseProcessor.SSOFailureResponse(ctx, "Your account is disabled, please contact support")
												} else {
													refreshID := generator.RefreshID()
													if !AccountProcessor.RecordReturningUser(refreshID, userID, state.RememberMe) {
														return ResponseProcessor.SSOFailureResponse(ctx, "Failed to login, please try again or contact support")
													} else {
														token := TokenProcessor.CreateFreshToken(
															userID,
															refreshID,
															UserModels.Find(AccountProcessor.GetUserType(userID)),
															state.RememberMe,
															fmt.Sprintf("%s-login", state.Provider),
														)
														ResponseProcessor.AttachAuthCookies(ctx, ResponseProcessor.GenerateAuthCookies(token))
														return ResponseProcessor.SSOSuccessResponse(ctx, token.AccessToken, state.FrontendState, state.Origin)
													}
												}
											} else {
												userID = generator.UserID()
												refreshID := generator.RefreshID()
												if !AccountProcessor.RecordNewUser(userID, "", user.Email, user.Name) {
													return ResponseProcessor.SSOFailureResponse(ctx, "Failed to create account, please contact support")
												} else {
													token := TokenProcessor.CreateFreshToken(
														userID,
														refreshID,
														UserModels.All.Viewer,
														state.RememberMe,
														fmt.Sprintf("%s-register", state.Provider),
													)
													ResponseProcessor.AttachAuthCookies(ctx, ResponseProcessor.GenerateAuthCookies(token))
													return ResponseProcessor.SSOSuccessResponse(ctx, token.AccessToken, state.FrontendState, state.Origin)
												}
											}
										}
									}
								}
							}
						}
					}

				}
			}
		}
	}
	return ResponseProcessor.SSOFailureResponse(ctx, "This didn't work, maybe try again??")
}
