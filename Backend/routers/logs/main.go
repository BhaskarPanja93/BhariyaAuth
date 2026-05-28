package logs

import (
	Middlewares "BhariyaAuth/middlewares"
	Logs "BhariyaAuth/processors/logs"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/logs/main", "", "Attaching Log Routes")

	SessionsRouter := APIGroup.Group("/logs")
	SessionsRouter.Get("/available", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Available)
	SessionsRouter.Get("/:date", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Date)
}
