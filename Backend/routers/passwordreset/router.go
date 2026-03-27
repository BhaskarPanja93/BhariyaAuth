package passwordreset

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	ForgotRouter := APIGroup.Group("/passwordreset")
	ForgotRouter.Post("/step1", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step1)
	ForgotRouter.Post("/step2", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step2)
}
