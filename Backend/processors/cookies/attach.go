package cookies

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	RequestProcessor "BhariyaAuth/processors/request"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenCombined) {

	start := RequestProcessor.GetRequestTime(ctx)

	Refresh := fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    token.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
		Path:     "/",
	}

	Csrf := fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    token.CSRF,
		HTTPOnly: false,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
		Path:     "/",
	}

	if token.RememberMe {
		Refresh.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Csrf.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Refresh.Expires = start.Add(Config.RefreshTokenExpireDelta)
		Csrf.Expires = start.Add(Config.RefreshTokenExpireDelta)
	} else {
		Refresh.MaxAge = 0
		Csrf.MaxAge = 0
	}

	ctx.Cookie(&Refresh)
	ctx.Cookie(&Csrf)
}

func AttachSSOCookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.SSOStateInCookie,
		Value:    value,
		MaxAge:   int(Config.SSOCookieExpireDelta.Seconds()),
		Expires:  time.Now().Add(Config.SSOCookieExpireDelta),
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   Config.Domain,
		Path:     "/",
	})
}

func AttachMFACookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    value,
		MaxAge:   0,
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
		Path:     "/",
	})
}
