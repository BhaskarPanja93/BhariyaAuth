package login

import (
	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	LoginRouter := authApp.Group("/login")

	LoginRouter.Post("/step1/:process", Step1)
	LoginRouter.Post("/step2", Step2)
}
