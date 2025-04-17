package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"mediafs/internal/service"
)

func BearerAuthMiddleware(auth *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") || !auth.CheckToken(strings.TrimPrefix(authHeader, "Bearer ")) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}
