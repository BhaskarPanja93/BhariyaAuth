package cookies

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"

	"github.com/gofiber/fiber/v3"
)

func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenCombinedT) {
	Refresh := fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    token.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	}
	Csrf := fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    token.CSRF,
		HTTPOnly: false,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	}
	if token.RememberMe {
		Refresh.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Csrf.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
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
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   Config.Domain,
	})

}

func AttachMFACookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    value,
		Secure:   false,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	})

}

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	})

}

func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.SSOStateInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   Config.Domain,
	})

}

func DetachMFACookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.Domain,
	})

}
