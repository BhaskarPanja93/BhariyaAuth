package cookies

import (
	Config "BhariyaAuth/constants/config"

	"github.com/gofiber/fiber/v3"
)

func DetachAuthCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.RefreshTokenInCookie)
	ctx.ClearCookie(Config.CSRFInCookie)
}

func DetachSSOCookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.SSOStateInCookie)
}

func DetachMFACookies(ctx fiber.Ctx) {
	ctx.ClearCookie(Config.MFATokenInCookie)
}
