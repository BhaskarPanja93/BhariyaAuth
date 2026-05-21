package token

import (
	TokenModels "BhariyaAuth/models/tokens"
	RequestProcessor "BhariyaAuth/processors/request"

	"github.com/gofiber/fiber/v3"
)

func VerifyCSRF(ctx fiber.Ctx, refresh TokenModels.RefreshToken) bool {

	if refresh.CSRF == "" {
		return false
	}

	cookie := ReadCookieCSRF(ctx)
	header := ReadHeaderCSRF(ctx)

	return cookie != "" &&
		header != "" &&
		cookie == refresh.CSRF &&
		header == refresh.CSRF
}

func AccessIsFresh(ctx fiber.Ctx, access TokenModels.AccessToken) bool {

	return RequestProcessor.GetRequestTime(ctx).Before(access.Expiry)
}

func RefreshIsFresh(ctx fiber.Ctx, refresh TokenModels.RefreshToken) bool {

	return RequestProcessor.GetRequestTime(ctx).Before(refresh.Expiry)
}
