package request

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

const requestTimeFlag = "request-time"

func SetRequestTime(ctx fiber.Ctx) time.Time {
	received := time.Now().UTC()
	ctx.Locals(requestTimeFlag, received)
	return received
}

func GetRequestTime(ctx fiber.Ctx) time.Time {
	if received, ok := ctx.Locals(requestTimeFlag).(time.Time); ok {
		return received
	}

	return time.Now().UTC()
}
