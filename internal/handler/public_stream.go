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
	"time"
)

func StreamPublicVideo(c *fiber.Ctx) error {
	path := c.Query("path")
	expiresStr := c.Query("expires")
	sig := c.Query("sig")

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "expired or invalid"})
	}

	if !util.VerifySignature(path, expires, sig) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	absPath, err := util.ResolveSafePath(mediafs.BaseDir, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fiber.ErrNotFound
	}
	defer file.Close()

	stat, _ := file.Stat()
	size := stat.Size()
	mime := service.DetectMimeType(absPath)

	rangeHeader := c.Get("Range")
	if rangeHeader == "" {
		c.Set("Content-Type", mime)
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
	c.Set("Content-Type", mime)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	c.Set("Content-Length", fmt.Sprintf("%d", length))
	c.Set("Accept-Ranges", "bytes")

	file.Seek(start, io.SeekStart)
	return c.SendStream(io.LimitReader(file, length))
}

func GeneratePublicLink(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}

	url := util.Generate24hPublicLink("http://localhost:8080", path)
	return c.JSON(fiber.Map{"url": url})
}
