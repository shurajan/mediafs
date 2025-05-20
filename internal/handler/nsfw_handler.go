package handler

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// GetNsfwFrameList - возвращает список JPEG-файлов из папки nsfw для конкретного видео
func GetNsfwFrameList(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем безопасное имя видеопапки
		videoname := filepath.Base(c.Params("videoname"))

		// Путь к папке nsfw
		nsfwDir := filepath.Join(baseDir, videoname, "nsfw")

		// Чтение содержимого директории
		entries, err := os.ReadDir(nsfwDir)
		if err != nil {
			log.Printf("Failed to read nsfw directory: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to read nsfw folder",
			})
		}

		// Составляем список только jpeg-файлов
		files := make([]string, 0)
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				ext := filepath.Ext(name)
				if ext == ".jpg" || ext == ".jpeg" {
					files = append(files, name)
				}
			}
		}

		return c.JSON(fiber.Map{
			"files": files,
		})
	}
}

// GetKeyFrameFile - простой обработчик, возвращающий кадр по имени файла
func GetNsfwFrameFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем имя видео из URL
		videoname := filepath.Base(c.Params("videoname"))

		// Получаем имя файла из URL
		filename := c.Params("filename")

		// Защита от path traversal
		filename = filepath.Base(filename)

		// Формируем путь к файлу ключевого кадра
		keyframePath := filepath.Join(baseDir, videoname, "nsfw", filename)

		// Проверяем, существует ли файл
		if _, err := os.Stat(keyframePath); os.IsNotExist(err) {
			log.Printf("Keyframe file not found: %s", keyframePath)
			return c.Status(404).JSON(fiber.Map{
				"error": "Keyframe not found",
			})
		}

		// Отправляем файл
		c.Set("Content-Type", "image/jpeg")
		c.Set("Cache-Control", "public, max-age=86400")
		return c.SendFile(keyframePath)
	}
}
