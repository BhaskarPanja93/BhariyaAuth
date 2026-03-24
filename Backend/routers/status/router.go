package status

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/gofiber/contrib/monitor"
)

func AttachRoutes(APIGroup fiber.Router) {
	StatusRouter := APIGroup.Group("/status")

	StatusRouter.Get("/metrics", monitor.New(monitor.Config{Title: "BhariyaAuth Metrics", Refresh: 1 * time.Second}))
	StatusRouter.All("/ping", func(ctx fiber.Ctx) error { return nil })
	StatusRouter.Get("/ip", func(ctx fiber.Ctx) error { return ctx.SendString(ctx.IP() + " <- " + strings.Join(ctx.IPs(), "/")) })
}
