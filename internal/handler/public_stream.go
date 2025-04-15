package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"log"
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

	log.Println("📥 StreamPublicVideo requested")
	log.Println("├─ Raw path (encoded):", path)
	log.Println("├─ Expires:", expiresStr)
	log.Println("├─ Signature provided:", sig)

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		log.Println("❌ Link expired or invalid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "expired or invalid"})
	}

	expectedSig := util.GenerateSignature(path, expires)
	isValid := util.VerifySignature(path, expires, sig)

	log.Println("├─ Signature expected:", expectedSig)
	log.Println("└─ Signature valid?", isValid)

	if !isValid {
		log.Println("❌ Invalid signature!")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	decodedPath, err := url.QueryUnescape(path)
	if err != nil {
		log.Println("❌ Failed to decode path:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid path encoding"})
	}

	absPath, err := util.ResolveSafePath(mediafs.BaseDir, decodedPath)
	if err != nil {
		log.Println("❌ Path resolution error:", err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("✅ Serving file:", absPath)

	file, err := os.Open(absPath)
	if err != nil {
		log.Println("❌ File open error:", err)
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
		c.Set("Content-Length", strconv.FormatInt(size, 10))
		c.Set("Accept-Ranges", "bytes")
		c.Context().SetStatusCode(fiber.StatusOK)
		_, err := io.Copy(c, file)
		if err != nil {
			log.Println("❌ io.Copy error:", err)
			return fiber.ErrInternalServerError
		}
		return nil
	}

	var start, end int64
	n, _ := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if n == 1 || end >= size {
		end = size - 1
	}
	length := end - start + 1

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	c.Status(fiber.StatusPartialContent)
	c.Set("Content-Type", mime)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	c.Set("Content-Length", strconv.FormatInt(length, 10))
	c.Set("Accept-Ranges", "bytes")

	_, err = io.CopyN(c, file, length)
	if err != nil {
		log.Println("❌ io.CopyN error:", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func GeneratePublicLink(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}

	baseURL, err := util.GetLocalBaseURL(8000)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resolve local IP"})
	}

	url := util.Generate24hPublicLink(baseURL, path)
	return c.JSON(fiber.Map{"url": url})
}
