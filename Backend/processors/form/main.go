package form

import (
	"github.com/gofiber/fiber/v3"
)

func ReadFormData(ctx fiber.Ctx, form any) bool {
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			return false
		}
	}
	return true
}
