package passwordreset

import (
	Middlewares "BhariyaAuth/middlewares"
	Logs "BhariyaAuth/processors/logs"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/passwordreset/main", "", "Attaching PasswordReset Routes")

	ForgotRouter := APIGroup.Group("/passwordreset")
	ForgotRouter.Post("/step1", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step1)
	ForgotRouter.Post("/step2", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Step2)
}
