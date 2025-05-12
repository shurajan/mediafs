package handler

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// GetKeyFrameFile - простой обработчик, возвращающий кадр по имени файла
func GetKeyFrameFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем имя видео из URL
		videoname := filepath.Base(c.Params("videoname"))

		// Получаем имя файла из URL
		filename := c.Params("filename")

		// Защита от path traversal
		filename = filepath.Base(filename)

		// Формируем путь к файлу ключевого кадра
		keyframePath := filepath.Join(baseDir, videoname, "keyframes", filename)

		// Проверяем, существует ли файл
		if _, err := os.Stat(keyframePath); os.IsNotExist(err) {
			log.Printf("Keyframe file not found: %s", keyframePath)
			return c.Status(404).JSON(fiber.Map{
				"error": "Keyframe not found",
			})
		}

		// Отправляем файл
		c.Set("Content-Type", "image/jpeg")
		c.Set("Cache-Control", "public, max-age=86400") // 24 часа кэширования
		return c.SendFile(keyframePath)
	}
}
