package handler

import (
	"encoding/base32"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mediafs/internal/media"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zeebo/blake3"
)

func ListFiles(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		type MediaFile struct {
			ID         string  `json:"id"`
			Name       string  `json:"name"`
			Size       int64   `json:"size"`
			Resolution *string `json:"resolution"`
			Duration   *int    `json:"duration"`
			PreviewURL *string `json:"previewURL"`
		}

		files := make([]MediaFile, 0)
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				if ext == ".mp4" {
					fullPath := filepath.Join(baseDir, entry.Name())
					info, err := os.Stat(fullPath)
					if err != nil {
						continue
					}

					resolution := media.GetVideoResolution(fullPath)
					duration := media.GetVideoDuration(fullPath)

					files = append(files, MediaFile{
						ID:         IDFromNameSize(entry.Name(), info.Size()),
						Name:       entry.Name(),
						Size:       info.Size(),
						Resolution: resolution,
						Duration:   duration,
						PreviewURL: nil,
					})
				}
			}
		}
		return c.JSON(files)
	}
}

func StreamFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename"))
		absPath := filepath.Join(baseDir, filename)

		file, err := os.Open(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		stat, err := file.Stat()
		if err != nil {
			file.Close()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		size := stat.Size()

		ext := filepath.Ext(absPath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		rangeHeader := c.Get("Range")
		if rangeHeader == "" {
			// Полная передача
			c.Status(fiber.StatusOK)
			c.Set("Content-Type", mimeType)
			c.Set("Accept-Ranges", "bytes")
			return StreamRange(c, file, 0, size-1, size)
		}

		// Обработка Range-запроса
		var start, end int64
		n, _ := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if n == 1 || end >= size {
			end = size - 1
		}
		if start > end || start >= size {
			file.Close()
			return c.Status(fiber.StatusRequestedRangeNotSatisfiable).JSON(fiber.Map{"error": "invalid range"})
		}

		c.Status(fiber.StatusPartialContent)
		c.Set("Content-Type", mimeType)
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		c.Set("Content-Length", strconv.FormatInt(end-start+1, 10))
		c.Set("Accept-Ranges", "bytes")

		return StreamRange(c, file, start, end, size)
	}
}

func DeleteFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename")) // Защита от ../
		fullPath := filepath.Join(baseDir, filename)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			return c.Status(404).JSON(fiber.Map{"error": "file not found"})
		}

		err = os.Remove(fullPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete"})
		}

		return c.JSON(fiber.Map{"message": "deleted"})
	}
}

func StreamRange(c *fiber.Ctx, file *os.File, start, end, totalSize int64) error {
	defer file.Close()

	length := end - start + 1

	_, err := file.Seek(start, io.SeekStart)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	buf := make([]byte, 512*1024) // ✅ 512 KB буфер
	var written int64
	for written < length {
		toRead := len(buf)
		remaining := length - written
		if remaining < int64(toRead) {
			toRead = int(remaining)
		}
		n, readErr := file.Read(buf[:toRead])
		if n > 0 {
			_, writeErr := c.Write(buf[:n])
			if writeErr != nil {
				fmt.Println("❌ Write error:", writeErr)
				return nil // клиент отключился — это нормально
			}
			written += int64(n)
		}
		if readErr != nil {
			if readErr != io.EOF {
				fmt.Println("❌ Read error:", readErr)
			}
			break
		}
	}

	return nil
}

func IDFromNameSize(name string, size int64) string {
	canonical := fmt.Sprintf("%s:%d", strings.ToLower(strings.TrimSpace(name)), size)

	hash := blake3.Sum256([]byte(canonical)) // 256‑бит
	sum := hash[:16]                         // берём первые 128‑бит

	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString(sum) // ≈26 символов
}
