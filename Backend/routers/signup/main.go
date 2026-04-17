package signup

import (
	Middlewares "BhariyaAuth/middlewares"
	Logs "BhariyaAuth/processors/logs"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/signup/main", "", "Attaching SignUp Routes")

	SignUpRouter := APIGroup.Group("/signup")
	SignUpRouter.Post("/step1", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step1)
	SignUpRouter.Post("/step2", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step2)
}
