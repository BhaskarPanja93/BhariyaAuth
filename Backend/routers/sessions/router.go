package sessions

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	SessionsRouter := authApp.Group("/sessions")

	SessionsRouter.Post("/fetch", Middlewares.RouteRateLimiter(20, 2*time.Minute, 3*time.Minute, 10*time.Minute), Fetch)
	SessionsRouter.Post("/revoke", Middlewares.RouteRateLimiter(20, 2*time.Minute, 3*time.Minute, 10*time.Minute), Revoke)
}
