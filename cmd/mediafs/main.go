package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"mediafs/internal/handler"
	"mediafs/internal/service"
)

func authMiddleware(c *fiber.Ctx) error {
	if service.IsTestMode() {
		if strings.HasPrefix(c.Get("Authorization"), "Bearer ") {
			return c.Next()
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized (test mode)"})
	}

	//TODO: - В будущем здесь может быть продовая проверка
	return c.Next()
}

func main() {
	if err := service.InitMediaFS(); err != nil {
		log.Fatalf("failed to initialize mediafs: %v", err)
	}

	if service.IsTestMode() {
		log.Println("Running in test mode")
		if err := service.InitTestEnv(); err != nil {
			log.Fatalf("failed to initialize test env: %v", err)
		}
	}

	app := fiber.New()
	app.Use(authMiddleware)

	app.Get("/files/list", handler.ListFiles)
	app.Post("/files/upload", handler.UploadFile)
	app.Get("/files/download", handler.DownloadFile)
	app.Delete("/files/delete", handler.DeleteFile)
	app.Put("/files/rename", handler.RenameFile)
	app.Post("/folders/create", handler.CreateFolder)

	log.Println("MediaFS running on :8080")
	log.Fatal(app.Listen(":8080"))
}
