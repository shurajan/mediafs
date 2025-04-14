package middleware

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/service"
	"os"
	"strings"
)

func AuthMiddleware(c *fiber.Ctx) error {
	if service.IsTestMode() {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		token := strings.TrimPrefix(auth, "Bearer ")

		expected, err := os.ReadFile(service.TestTokenFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token read error"})
		}
		if token != strings.TrimSpace(string(expected)) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		return c.Next()
	}

	// TODO: Add production logic here
	return c.Next()
}
