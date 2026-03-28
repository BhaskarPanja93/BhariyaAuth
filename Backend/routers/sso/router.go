package sso

import (
	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(APIGroup fiber.Router) {
	SSORouter := APIGroup.Group("/sso")
	SSORouter.Get("/:"+ProviderParam, Step1)
	SSORouter.Get("/callback/:"+ProviderParam, Step2)
}
