package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/service"
)

// ChunkUploadHandler handles uploading a file in chunks via service layer.
func ChunkUploadHandler(c *fiber.Ctx) error {
	path := c.Query("path")
	chunkIndex := c.Query("index")
	isLast := c.Query("last") == "1"

	if path == "" || chunkIndex == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path and index query required"})
	}

	fileHeader, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "chunk file missing"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open chunk"})
	}
	defer src.Close()

	if err := service.AppendChunk(path, src); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if isLast {
		if err := service.FinalizeChunkUpload(path); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("✅ Finalized upload for %s\n", path)
		return c.JSON(fiber.Map{"message": "upload complete"})
	}

	fmt.Printf("📦 Chunk received for %s (index %s)\n", path, chunkIndex)
	return c.JSON(fiber.Map{"message": "chunk received"})
}
