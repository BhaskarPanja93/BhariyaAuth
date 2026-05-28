package request

import "github.com/gofiber/fiber/v3"

const rateLimitFlag = "rate-limit-weight"

func AddRateLimitWeight(ctx fiber.Ctx, value uint32) uint32 {
	var updated uint32
	var old uint32
	var ok bool
	if old, ok = ctx.Locals(rateLimitFlag).(uint32); !ok {
		old = 1
	}
	updated = old + value
	ctx.Locals(rateLimitFlag, updated)
	return updated
}

func GetRateLimitWeight(ctx fiber.Ctx) uint32 {
	if old, ok := ctx.Locals(rateLimitFlag).(uint32); ok {
		return old
	}
	return 1
}
