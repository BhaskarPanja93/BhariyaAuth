package middlewares

import (
	Logger "BhariyaAuth/processors/logs"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

var (
	MaxResponseTimeAllowed = 4 * time.Second
)

func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now().UTC()
		err := ctx.Next()
		end := time.Now().UTC()
		dur := end.Sub(start)
		if dur > MaxResponseTimeAllowed {
			Logger.AccidentalFailure(fmt.Sprintf("[Profiling] Server took %v seconds for request completion", dur))
		}
		ctx.Set("Requested-Started", fmt.Sprintf("%v", start))
		ctx.Set("Time-Taken", fmt.Sprintf("%v", dur))
		return err
	}
}
