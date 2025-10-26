package register

import (
	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	RegisterRouter := authApp.Group("/register")

	RegisterRouter.Post("/step1", Step1)
	RegisterRouter.Post("/step2", Step2)
}
