package request

import (
	StringProcessor "BhariyaAuth/processors/string"

	"github.com/gofiber/fiber/v3"
)

// requestIdFlag is the context key used to store per-request rate limit weight.
const requestIdFlag = "request-id"

// SetRequestId sets a request ID for the request.
//
// Behavior:
// - Holds the same requestID for the entirety of the request.
func SetRequestId(ctx fiber.Ctx) string {

	id := StringProcessor.SafeString(6)
	ctx.Locals(requestIdFlag, id)
	return id
}

// GetRequestId retrieves the requestID for the request.
//
// Returns:
// - string requestID of the request.
func GetRequestId(ctx fiber.Ctx) string {

	return ctx.Locals(requestIdFlag).(string)
}
