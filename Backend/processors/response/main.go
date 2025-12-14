package response

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func SSOSuccessPopup(ctx fiber.Ctx, token string) error {
	return ctx.Type("html").SendString(fmt.Sprintf(`
<html>
<head>
<script>
function onAuthSuccess() {
    window._authSuccess = true;
    window.opener?.postMessage({ success: true, token: '%s'}, window.location.origin);
    window.close();
}
window.addEventListener('beforeunload', () => {
    if (!window._authSuccess) {
        window.opener?.postMessage({ success: false }, window.location.origin);
    }
});
</script>
</head>
<body onload="onAuthSuccess()">
<h2>That was easy!</h2>
</body>
</html>
`, token))
}

func SSOFailurePopup(ctx fiber.Ctx, reason string) error {
	return ctx.Type("html").SendString(fmt.Sprintf(`
<html>
	<script>
		window.addEventListener('beforeunload', () => {
			if (!window._authSuccess) {
				window.opener?.postMessage({ success: false }, window.location.origin);
			}
		})
	</script>
	<body>
		<h2>%s</h2>
	</body>
</html>
`, reason))
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
