package response

import (
	Important "BhariyaAuth/constants/config"
	CookieModels "BhariyaAuth/models/cookies"
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

func GenerateAuthCookies(token TokenModels.NewTokenT) CookieModels.ResponseCookiesT {
	return CookieModels.ResponseCookiesT{
		Refresh: fiber.Cookie{
			Name:     Important.RefreshTokenInCookie,
			Value:    token.RefreshToken,
			Domain:   Important.CookieDomain,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteStrictMode,
		},
		Csrf: fiber.Cookie{
			Name:     Important.CSRFInCookie,
			Value:    token.CSRF,
			Domain:   Important.CookieDomain,
			HTTPOnly: false,
			Secure:   true,
			SameSite: fiber.CookieSameSiteStrictMode,
		},
	}
}

func SSOSuccessResponse(ctx fiber.Ctx, token string, state string, origin string) error {
	return ctx.Type("html").SendString(fmt.Sprintf(`
<html>
<head>
<script>
function onAuthSuccess(value) {
    window._authSuccess = true;
    window.opener?.postMessage({ type: 'SSO_SUCCESS', token: value, state: '%s'}, '%s');
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

func AttachAuthCookies(ctx fiber.Ctx, cookies CookieModels.ResponseCookiesT) {
	ctx.Cookie(&cookies.Refresh)
	ctx.Cookie(&cookies.Csrf)
}

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Important.RefreshTokenInCookie)
	ctx.ClearCookie(Important.CSRFInCookie)
}
