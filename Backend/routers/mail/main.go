package mail

import (
	Middlewares "BhariyaAuth/middlewares"
	Logs "BhariyaAuth/processors/logs"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/mail/main", "", "Attaching Mail Routes")

	MailRouter := APIGroup.Group("/mail")
	MailRouter.Post("/send", Middlewares.RouteRateLimiter(600_000, time.Minute, 2*time.Minute), Sender)
}
