package middlewares

import (
	Logger "BhariyaAuth/processors/logs"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()
		dur := time.Since(start)
		if dur.Seconds() > 2 {
			Logger.AccidentalFailure(fmt.Sprintf("[Profiling] Server took %v (>2) seconds for request", dur))
		}
		ctx.Set("Requested-Started", fmt.Sprintf("%v", start))
		ctx.Set("Time-Taken", fmt.Sprintf("%v", dur))
		return err
	}
}
