package ratelimit

import "github.com/gofiber/fiber/v3"

// flagKey is the context key used to store per-request rate limit weight.
const flagKey = "rate-limit-weight"

// Init initializes the rate limit weight for the current request.
//
// - Sets a base weight for the request.
// - This weight represents the "cost" of the request in the rate limiter.
//
// Usage:
// - Must be called at the beginning of request processing (middleware).
//
// Behavior:
// - Default weight is 1 (normal request cost).
func Init(ctx fiber.Ctx) {

	ctx.Locals(flagKey, uint32(1))
}

// Add increases the rate limit weight for the current request.
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
func Add(ctx fiber.Ctx, value uint32) {

	current := ctx.Locals(flagKey).(uint32)
	ctx.Locals(flagKey, current+value)
}

// Get retrieves the current rate limit weight for the request.
//
// - Used by rate limiter middleware to determine how much to increment counters.
//
// Returns:
// - uint32 weight of the request.
//
// Behavior:
// - Defaults to 1 if not initialized (defensive fallback).
func Get(ctx fiber.Ctx) uint32 {

	return ctx.Locals(flagKey).(uint32)
}
