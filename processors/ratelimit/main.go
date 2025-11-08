package ratelimit

import "github.com/gofiber/fiber/v3"

const flagKey = "RateLimitValue"

func SetValue(ctx fiber.Ctx) {
	ctx.Locals(flagKey, true)
}
func CheckValue(ctx fiber.Ctx) uint16 {
	if ctx.Locals("CountTowardsRateLimit") == true {
		return 100
	} else {
		return 1
	}
}
