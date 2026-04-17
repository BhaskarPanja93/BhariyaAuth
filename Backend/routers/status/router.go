package status

import (
	Logs "BhariyaAuth/processors/logs"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/gofiber/contrib/monitor"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/status/main", "", "Attaching Status Routes")

	StatusRouter := APIGroup.Group("/status")
	StatusRouter.All("/ready", func(ctx fiber.Ctx) error { return nil })
	StatusRouter.Get("/metrics", monitor.New(monitor.Config{Title: "BhariyaAuth Metrics", Refresh: 1 * time.Second})) // TODO: replace with custom solution
	StatusRouter.Get("/ip", func(ctx fiber.Ctx) error { return ctx.SendString(ctx.IP() + " <- " + strings.Join(ctx.IPs(), "/")) })
}
