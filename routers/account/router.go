package account

import (
	"github.com/gofiber/fiber/v3"
)

func AttachRoutes(authApp fiber.Router) {
	AccountRouter := authApp.Group("/account")

	AccountRouter.Post("/refresh", ProcessRefresh)
	AccountRouter.Post("/logout", ProcessLogout)
	AccountRouter.All("/me", Me)
}
