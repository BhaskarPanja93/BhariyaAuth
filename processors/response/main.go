package response

import (
	Config "BhariyaAuth/constants/config"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func CombineResponses(
	reply ResponseModels.GeneralT,
	auth ResponseModels.AuthT,
	notifications []string,
	secret map[string]interface{},
	extra map[string]interface{},
) ResponseModels.CombinedT {
	return ResponseModels.CombinedT{
		Reply:         reply,
		Notifications: notifications,
		Auth:          auth,
		Secret:        secret,
		Extra:         extra,
	}
}

func SSOSuccessResponse(ctx fiber.Ctx, token string, state string, origin string) error {
	return ctx.Type("html").SendString(fmt.Sprintf(`
<html>
<head>
<script>
function onAuthSuccess(value) {
    window._authSuccess = true;
    window.opener?.postMessage({ type: 'SSO_SUCCESS', token: 'value', state: '%s'}, '%s');
    window.close();
}
window.addEventListener('beforeunload', () => {
    if (!window._authSuccess) {
        window.opener?.postMessage({ type: 'SSO_CLOSED', state: '%s'}, '%s');
    }
});
</script>
</head>
<body onload="onAuthSuccess('%s')">
<h2>That was easy!</h2>
</body>
</html>
`, state, origin, state, origin, token))
}

func SSOFailureResponse(ctx fiber.Ctx, reason string) error {
	return ctx.Type("html").SendString(fmt.Sprintf("<html><body><h2>%s</h2></body></html>", reason))
}

func AttachAuthCookies(ctx fiber.Ctx, token TokenModels.NewTokenT) {
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
		Csrf.MaxAge = int(Config.RefreshTokenExpireDelta.Seconds())
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

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: Config.RefreshTokenInCookie, Value: ""})
	ctx.Cookie(&fiber.Cookie{Name: Config.CSRFInCookie, Value: ""})
}

func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: Config.SSOStateInCookie, Value: ""})
}
