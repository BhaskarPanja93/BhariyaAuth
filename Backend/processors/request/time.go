package request

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// requestTimeFlag is the context key used to store per-request rate limit weight.
const requestTimeFlag = "request-time"

// SetRequestTime sets the time for request arrival.
//
// Behavior:
// - Holds the same requestID for the entirety of the request.
func SetRequestTime(ctx fiber.Ctx) time.Time {

	received := time.Now().UTC()
	ctx.Locals(requestTimeFlag, received)
	return received
}

// GetRequestTime retrieves the time for the request arrival.
//
// Returns:
// - string requestID of the request.
func GetRequestTime(ctx fiber.Ctx) time.Time {

	return ctx.Locals(requestTimeFlag).(time.Time)
}
