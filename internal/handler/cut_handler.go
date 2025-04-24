package handler

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/service"
)

type CutRequest struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Name string `json:"name"`
}

func CutHandler(cut *service.CutService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := c.Params("videoname")

		var req CutRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid json")
		}
		if req.From < 0 || req.To <= req.From {
			return fiber.NewError(fiber.StatusBadRequest, "invalid from/to range")
		}

		clipName, err := cut.CreateClip(filename, req.From, req.To, req.Name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"message": "cut created",
			"file":    clipName,
			"url":     "/videos/" + filename + "/" + clipName,
		})
	}
}
