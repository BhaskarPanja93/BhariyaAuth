package passwordreset

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	ForgotRouter := authApp.Group("/passwordreset")

	ForgotRouter.Post("/step2", Middlewares.RouteRateLimiter(10, time.Minute, 3*time.Minute, 10*time.Minute), Step2)
	ForgotRouter.Post("/step1", Middlewares.RouteRateLimiter(10, time.Minute, 3*time.Minute, 10*time.Minute), Step1)
}
