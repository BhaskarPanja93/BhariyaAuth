package account

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	AccountRouter := APIGroup.Group("/account")
	AccountRouter.Post("/logout", Logout)
	AccountRouter.Post("/refresh", Middlewares.RouteRateLimiter(60_000, time.Minute, 2*time.Minute), Refresh)
}
