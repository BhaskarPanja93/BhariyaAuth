package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// LocalsMiddleware is an App based Middleware that must be attached at root level as this provides the "request-start" value to ctx
func LocalsMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		ctx.Set("Reached-API", "yes")
		ctx.Locals("request-start", time.Now().UTC())
		err := ctx.Next()
		return err
	}
}
