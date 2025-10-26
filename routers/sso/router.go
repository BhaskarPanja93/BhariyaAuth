package sso

import "github.com/gofiber/fiber/v3"

func AttachRoutes(authApp fiber.Router) {
	SSORouter := authApp.Group("/sso")

	SSORouter.Get("/:processor", Step1)
	SSORouter.Get("/callback", Step2)
}
