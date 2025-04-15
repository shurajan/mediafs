package main

import (
	"log"
	"mediafs/internal/mediafs"
	"mediafs/internal/router"

	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := mediafs.Init(); err != nil {
		log.Fatalf("failed to initialize mediafs: %v", err)
	}

	app := fiber.New()

	// Регистрируем все маршруты
	router.RegisterRoutes(app)

	log.Println("MediaFS running on :8080")
	log.Fatal(app.Listen(":8080"))
}
