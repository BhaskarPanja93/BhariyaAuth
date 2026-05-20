package form

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

func ReadFormData(ctx fiber.Ctx, form any) error {

	if err := ctx.Bind().Form(form); err == nil {
		return nil
	}

	if err := ctx.Bind().Body(form); err == nil {
		return nil
	}

	return errors.New("form could not be read")
}
