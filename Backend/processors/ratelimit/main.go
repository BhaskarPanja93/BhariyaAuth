package ratelimit

import "github.com/gofiber/fiber/v3"

const flagKey = "rate-limit-value"

func Set(ctx fiber.Ctx) {
	ctx.Locals(flagKey, 1)
}
func Add(ctx fiber.Ctx, value uint32) {
	ctx.Locals(flagKey, ctx.Locals(flagKey).(uint32)+value)
}
func Get(ctx fiber.Ctx) uint32 {
	return ctx.Locals(flagKey).(uint32)
}
