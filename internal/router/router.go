package router

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/handler"
	"mediafs/internal/middleware"
)

func RegisterRoutes(app *fiber.App) {
	// Приватные маршруты (требуют Authorization)
	auth := app.Group("/", middleware.AuthMiddleware)

	auth.Get("/files/list", handler.ListFiles)
	auth.Post("/files/upload", handler.UploadFile)
	auth.Get("/files/download", handler.DownloadFile)
	auth.Delete("/files/delete", handler.DeleteFile)
	auth.Put("/files/rename", handler.RenameFile)
	auth.Post("/folders/create", handler.CreateFolder)
	auth.Get("/api/media/public/share", handler.GeneratePublicLink)
	app.Get("/api/media/public/stream", handler.StreamPublicVideo)
}
