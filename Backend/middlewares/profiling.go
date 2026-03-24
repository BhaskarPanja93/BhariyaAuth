package middlewares

import (
	Logger "BhariyaAuth/processors/logs"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	MaxResponseTimeAllowed = 4 * time.Second
)

// ProfilingMiddleware is an App based Middleware that attaches server response times to client header for better debugging on slow routes
func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := ctx.Locals("request-start").(time.Time)
		err := ctx.Next()
		end := time.Now().UTC()
		dur := end.Sub(start)
		// Also Log if server took too long to respond
		if dur > MaxResponseTimeAllowed {
			Logger.AccidentalFailure(fmt.Sprintf("[Profiling] Server took %v seconds for request completion", dur))
		}
		ctx.Set("Requested-Started", fmt.Sprintf("%v", start))
		ctx.Set("Time-Taken", fmt.Sprintf("%v", dur))
		return err
	}
}
