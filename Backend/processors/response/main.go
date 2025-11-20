package response

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func SSOSuccessPopup(ctx fiber.Ctx, token string, origin string) error {
	return ctx.Type("html").SendString(fmt.Sprintf(`
<html>
<head>
<script>
function onAuthSuccess() {
    window._authSuccess = true;
    window.opener?.postMessage({ type: 'SUCCESS', token: '%s'}, '%s');
    window.close();
}
window.addEventListener('beforeunload', () => {
    if (!window._authSuccess) {
        window.opener?.postMessage({ type: 'CLOSED'}, '%s');
    }
});
</script>
</head>
<body onload="onAuthSuccess()">
<h2>That was easy!</h2>
</body>
</html>
`, token, origin, origin))
}

func SSOFailurePopup(ctx fiber.Ctx, reason string) error {
	return ctx.Type("html").SendString(fmt.Sprintf("<html><body><h2>%s</h2></body></html>", reason))
}

func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenCombinedT) {
	Refresh := fiber.Cookie{
		Name:     Config.RefreshTokenInCookie,
		Value:    token.RefreshToken,
		Domain:   Config.CookieDomain,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
	}
	Csrf := fiber.Cookie{
		Name:     Config.CSRFInCookie,
		Value:    token.CSRF,
		Domain:   Config.CookieDomain,
		HTTPOnly: false,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
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
		Domain:   Config.CookieDomain,
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func AttachMFACookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     Config.MFATokenInCookie,
		Value:    value,
		Domain:   Config.CookieDomain,
		Secure:   false,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteNoneMode,
	})
}

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: Config.RefreshTokenInCookie, Value: "", MaxAge: 1})
	ctx.Cookie(&fiber.Cookie{Name: Config.CSRFInCookie, Value: "", MaxAge: 1})
}

func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: Config.SSOStateInCookie, Value: "", MaxAge: 1})
}

func DetachMFACookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: Config.MFATokenInCookie, Value: "", MaxAge: 1})
}
