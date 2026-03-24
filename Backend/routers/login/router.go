package login

import (
	Middlewares "BhariyaAuth/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	LoginRouter := APIGroup.Group("/login")
	LoginRouter.Post("/step1/:"+ProcessParam, Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step1)
	LoginRouter.Post("/step2", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step2)
}
