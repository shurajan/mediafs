package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mediafs/internal/mediafs"
	"mediafs/internal/service"
	"mediafs/internal/util"
	"os"
	"strconv"
)

func StreamVideo(c *fiber.Ctx) error {
	relPath := c.Query("path")
	if relPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing file path"})
	}

	absPath, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	size := stat.Size()

	mimeType := service.DetectMimeType(absPath)
	rangeHeader := c.Get("Range")

	if rangeHeader == "" {
		c.Set("Content-Type", mimeType)
		c.Set("Content-Length", fmt.Sprintf("%d", size))
		c.Set("Accept-Ranges", "bytes")
		return c.SendStream(file)
	}

	var start, end int64
	fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if end == 0 || end >= size {
		end = size - 1
	}
	length := end - start + 1

	c.Status(fiber.StatusPartialContent)
	c.Set("Content-Type", mimeType)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	c.Set("Content-Length", strconv.FormatInt(length, 10))
	c.Set("Accept-Ranges", "bytes")

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.SendStream(io.LimitReader(file, length))
}
