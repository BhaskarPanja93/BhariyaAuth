package login

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	LoginRouter := authApp.Group("/login")

	LoginRouter.Post("/step2", Middlewares.RouteRateLimiter(10, time.Minute, 3*time.Minute, 10*time.Minute), Step2)
	LoginRouter.Post("/step1/:process", Middlewares.RouteRateLimiter(10, time.Minute, 3*time.Minute, 10*time.Minute), Step1)
}
