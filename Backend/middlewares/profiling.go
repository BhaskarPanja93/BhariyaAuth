package middlewares

import (
	Logger "BhariyaAuth/processors/logs"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now().UTC()
		ctx.Locals("request-start", start)
		err := ctx.Next()
		end := time.Now().UTC()
		dur := end.Sub(start)
		if dur.Seconds() > 4 {
			Logger.AccidentalFailure(fmt.Sprintf("[Profiling] Server took %v (>4) seconds for request completion", dur))
		}
		ctx.Set("Requested-Started", fmt.Sprintf("%v", start))
		ctx.Set("Time-Taken", fmt.Sprintf("%v", dur))
		return err
	}
}
