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

func GenerateCookies(token TokenModels.NewTokenT) CookieModels.ResponseCookiesT {
	return CookieModels.ResponseCookiesT{
		Refresh: fiber.Cookie{
			Name:     Important.RefreshTokenInCookie,
			Value:    token.RefreshToken,
			Domain:   Important.CookieDomain,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		},
		Csrf: fiber.Cookie{
			Name:     Important.CSRFInCookie,
			Value:    token.CSRF,
			Domain:   Important.CookieDomain,
			HTTPOnly: false,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		},
	}
}

func SSOSuccessResponse(newToken TokenModels.NewTokenT, origin string) string {
	return fmt.Sprintf(`
<html>
<head>
<script>
window.addEventListener('beforeunload', () => {
    if (!window._authSuccess) {
        window.opener?.postMessage({ type: 'SSO_CLOSED' }, '%s');
    }
});
function onAuthSuccess(token) {
    window._authSuccess = true;
    window.opener?.postMessage({ type: 'SSO_SUCCESS', auth_token: token }, '%s');
    window.close();
}
</script>
</head>
<body onload="onAuthSuccess('%s')">
<h2>That was easy!</h2>
</body>
</html>
`, origin, origin, newToken.AccessToken)
}

func SSOFailureResponse() string {
	return `
<html>
<body>
<h2>This didn't work, maybe try again!</h2>
</body>
</html>
`
}

func AttachCookies(ctx fiber.Ctx, cookies CookieModels.ResponseCookiesT) {
	ctx.Cookie(&cookies.Refresh)
	ctx.Cookie(&cookies.Csrf)
}

func DetachCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Important.RefreshTokenInCookie)
	ctx.ClearCookie(Important.CSRFInCookie)
}
