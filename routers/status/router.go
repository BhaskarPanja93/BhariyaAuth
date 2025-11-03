package status

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/gofiber/contrib/monitor"
)

func AttachRoutes(authApp fiber.Router) {
	StatusRouter := authApp.Group("/status")

	StatusRouter.Get("/metrics", monitor.New(monitor.Config{Title: "Bhariya Auth Metrics", Refresh: 1 * time.Second}))
	StatusRouter.All("/ping", func(ctx fiber.Ctx) error {
		return nil
	})
}
