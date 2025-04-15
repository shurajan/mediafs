package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mediafs/internal/mediafs"
	"mediafs/internal/service"
	"mediafs/internal/util"
	"net/url"
	"os"
	"strconv"
	"time"
)

func StreamPublicVideo(c *fiber.Ctx) error {
	path := c.Query("path")
	expiresStr := c.Query("expires")
	sig := c.Query("sig")

	// Лог
	fmt.Println("📥 StreamPublicVideo requested")
	fmt.Println("├─ Raw path (encoded): ", path)
	fmt.Println("├─ Expires:            ", expiresStr)
	fmt.Println("├─ Signature provided: ", sig)

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		fmt.Println("❌ Link expired or invalid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "expired or invalid"})
	}

	expectedSig := util.GenerateSignature(path, expires)
	isValid := util.VerifySignature(path, expires, sig)

	fmt.Println("├─ Signature expected: ", expectedSig)
	fmt.Println("└─ Signature valid?    ", isValid)

	if !isValid {
		fmt.Println("❌ Invalid signature!")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	decodedPath, err := url.QueryUnescape(path)
	if err != nil {
		fmt.Println("❌ Failed to decode path:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid path encoding"})
	}

	absPath, err := util.ResolveSafePath(mediafs.BaseDir, decodedPath)
	if err != nil {
		fmt.Println("❌ Path resolution error:", err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	fmt.Println("✅ Serving file:", absPath)

	file, err := os.Open(absPath)
	if err != nil {
		fmt.Println("❌ File open error:", err)
		return fiber.ErrNotFound
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	size := stat.Size()
	mime := service.DetectMimeType(absPath)

	rangeHeader := c.Get("Range")
	if rangeHeader == "" {
		c.Set("Content-Type", mime)
		c.Set("Content-Length", fmt.Sprintf("%d", size))
		c.Set("Accept-Ranges", "bytes")

		c.Context().SetStatusCode(fiber.StatusOK)
		_, err := io.Copy(c, file)
		if err != nil {
			fmt.Println("❌ io.Copy error:", err)
			return fiber.ErrInternalServerError
		}
		return nil
	}

	// Обработка Range
	var start, end int64
	n, _ := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if n == 1 || end >= size {
		end = size - 1
	}
	length := end - start + 1

	c.Status(fiber.StatusPartialContent)
	c.Set("Content-Type", mime)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	c.Set("Content-Length", fmt.Sprintf("%d", length))
	c.Set("Accept-Ranges", "bytes")

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	_, err = io.CopyN(c, file, length)
	if err != nil {
		fmt.Println("❌ io.CopyN error:", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func GeneratePublicLink(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}

	url := util.Generate24hPublicLink("http://localhost:8080", path)
	return c.JSON(fiber.Map{"url": url})
}
