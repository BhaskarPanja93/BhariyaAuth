package access

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	AccountRouter := APIGroup.Group("/access")
	AccountRouter.Post("/logout", Logout)
	AccountRouter.Post("/refresh", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Refresh)
}
