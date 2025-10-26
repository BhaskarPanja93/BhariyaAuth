package middlewares

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func ProfilingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		c.Set("Requested-Started", fmt.Sprintf("%v", start))
		c.Set("Time-Taken", fmt.Sprintf("%v", duration))
		return err
	}
}
