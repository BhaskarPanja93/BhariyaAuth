package account

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	AccountRouter := authApp.Group("/account")

	AccountRouter.Post("/logout", ProcessLogout)
	AccountRouter.Post("/refresh", Middlewares.RouteRateLimiter(60, time.Minute, 3*time.Minute, 10*time.Minute), ProcessRefresh)
}
