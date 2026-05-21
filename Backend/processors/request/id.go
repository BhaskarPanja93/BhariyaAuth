package request

import (
	StringProcessor "BhariyaAuth/processors/string"

	"github.com/gofiber/fiber/v3"
)

const requestIdFlag = "request-id"

func SetRequestId(ctx fiber.Ctx) string {
	if existing, ok := ctx.Locals(requestIdFlag).(string); ok && existing != "" {
		return existing
	}

	id := StringProcessor.SafeString(6)
	ctx.Locals(requestIdFlag, id)
	return id
}

func GetRequestId(ctx fiber.Ctx) string {
	if existing, ok := ctx.Locals(requestIdFlag).(string); ok && existing != "" {
		return existing
	}

	return SetRequestId(ctx)
}
