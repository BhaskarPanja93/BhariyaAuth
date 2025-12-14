package ratelimit

import "github.com/gofiber/fiber/v3"

const flagKey = "RateLimitValue"

func Set(ctx fiber.Ctx) {
	ctx.Locals(flagKey, true)
}
func Get(ctx fiber.Ctx) uint16 {
	if ctx.Locals(flagKey) == true {
		return 100
	}
	return 1
}
