package handler

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/service"
)

func AuthHandler(auth *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid json")
		}

		if !auth.CheckPassword(req.Password) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid password")
		}

		token := auth.GenerateToken()
		return c.JSON(fiber.Map{"token": token})
	}
}
