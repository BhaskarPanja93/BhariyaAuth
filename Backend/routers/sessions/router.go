package sessions

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	SessionsRouter := APIGroup.Group("/sessions")
	SessionsRouter.Post("/fetch", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Fetch)
	SessionsRouter.Post("/revoke", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Revoke)
}
