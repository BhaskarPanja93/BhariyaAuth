package request

import "github.com/gofiber/fiber/v3"

// rateLimitFlag is the context key used to store per-request rate limit weight.
const rateLimitFlag = "rate-limit-weight"

// AddRateLimitWeight increases the rate limit weight for the current request.
//
// - Allows dynamic adjustment of request cost.
// - Typically used when:
//   - validation fails
//   - suspicious behavior detected
//   - expensive operations performed
//
// Parameters:
// - value: additional weight to add.
//
// Behavior:
// - Accumulates weight across the request lifecycle.
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

// GetRateLimitWeight retrieves the current rate limit weight for the request.
//
// - Used by rate limiter middleware to determine how much to increment counters.
//
// Returns:
// - uint32 weight of the request.
//
// Behavior:
// - Defaults to 1 if not initialized (defensive fallback).
func GetRateLimitWeight(ctx fiber.Ctx) uint32 {
	if old, ok := ctx.Locals(rateLimitFlag).(uint32); ok {
		return old
	}
	return 1
}
