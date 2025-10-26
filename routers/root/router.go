package root

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/gofiber/contrib/monitor"
)

func AttachRoutes(authApp fiber.Router) {
	authApp.Get("/metrics", monitor.New(monitor.Config{Title: "Bhariya Auth Metrics", Refresh: 1 * time.Second}))
	authApp.All("/ping", func(ctx fiber.Ctx) error {
		return nil
	})
	authApp.Post("/refresh", ProcessRefresh)
	authApp.Post("/logout", ProcessLogout)
	authApp.All("/me", Me)
}
