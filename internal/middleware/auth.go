package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	const expectedToken = "Bearer test-token-123"

	//TODO  - добавить авторизацию
	if c.Get("Authorization") != expectedToken {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	return c.Next()
}
