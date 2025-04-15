package handler

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ListFiles(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		type MediaFile struct {
			Name       string  `json:"name"`
			Size       int64   `json:"size"`
			Resolution *string `json:"resolution"` // Пока null
			PreviewURL *string `json:"previewURL"` // Пока null
		}

		files := make([]MediaFile, 0)
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				if ext == ".mp4" || ext == ".ts" {
					fullPath := filepath.Join(baseDir, entry.Name())
					info, err := os.Stat(fullPath)
					if err != nil {
						continue
					}

					files = append(files, MediaFile{
						Name:       entry.Name(),
						Size:       info.Size(),
						Resolution: nil, // пока не определяем
						PreviewURL: nil, // можно позже формировать
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
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
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
			c.Set("Content-Type", mimeType)
			c.Set("Content-Length", fmt.Sprintf("%d", size))
			c.Set("Accept-Ranges", "bytes")
			c.Status(fiber.StatusOK)

			_, err := io.Copy(c, file)
			return err
		}

		// Обработка Range-запроса
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
		c.Set("Content-Type", mimeType)
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		c.Set("Content-Length", strconv.FormatInt(length, 10))
		c.Set("Accept-Ranges", "bytes")

		_, err = io.CopyN(c, file, length)
		if err != nil {
			fmt.Println("❌ io.CopyN error:", err)
			return fiber.ErrInternalServerError
		}
		return nil
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
