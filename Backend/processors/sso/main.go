package sso

import (
	HTMLTemplates "BhariyaAuth/models/html"
	"time"

	"github.com/gofiber/fiber/v3"
)

func SuccessPopup(ctx fiber.Ctx, token string, expires time.Time, state string) error {

	return ctx.
		Type("html").
		SendString(
			HTMLTemplates.SSOSuccessPopup(
				token,
				expires.Format(time.RFC3339),
				state,
			),
		)
}

func FailurePopup(ctx fiber.Ctx, reason string) error {

	return ctx.
		Type("html").
		SendString(
			HTMLTemplates.SSOFailurePopup(reason),
		)
}
