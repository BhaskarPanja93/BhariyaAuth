package access

import (
	Middlewares "BhariyaAuth/middlewares"
	Logs "BhariyaAuth/processors/logs"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/access/main", "", "Attaching Access Routes")

	AccountRouter := APIGroup.Group("/access")
	AccountRouter.Post("/logout", Logout)
	AccountRouter.Post("/refresh", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Refresh)
}
