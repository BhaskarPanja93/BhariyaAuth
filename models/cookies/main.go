package cookies

import (
	"github.com/gofiber/fiber/v3"
)

type ResponseCookiesT struct {
	Refresh fiber.Cookie
	Csrf    fiber.Cookie
}
