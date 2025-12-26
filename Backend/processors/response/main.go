package response

import (
	Config "BhariyaAuth/constants/config"
	HTMLTemplates "BhariyaAuth/models/html"
	TokenModels "BhariyaAuth/models/tokens"

	"github.com/gofiber/fiber/v3"
)

func SSOSuccessPopup(ctx fiber.Ctx, token string) error {
	return ctx.Type("html").SendString(HTMLTemplates.SSOSuccessPopup(token))
}

func SSOFailurePopup(ctx fiber.Ctx, reason string) error {
	return ctx.Type("html").SendString(HTMLTemplates.SSOFailurePopup(reason))
}

func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenCombinedT) {
	Refresh := fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    token.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	}
	Csrf := fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    token.CSRF,
		HTTPOnly: false,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	}
	if token.RememberMe {
		Refresh.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
		Csrf.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
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
		Domain:   Config.CookieDomain,
	})

}

func AttachMFACookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    value,
		Secure:   false,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	})

}

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	})

}

func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.SSOStateInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   Config.CookieDomain,
	})

}

func DetachMFACookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    "",
		MaxAge:   1,
		SameSite: fiber.CookieSameSiteStrictMode,
		Domain:   Config.CookieDomain,
	})

}
