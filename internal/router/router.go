package router

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/handler"
	"mediafs/internal/middleware"
)

func RegisterRoutes(app *fiber.App) {
	registerAuthRoutes(app)
	registerPublicRoutes(app)
}

func registerAuthRoutes(app *fiber.App) {
	auth := app.Group("/api", middleware.AuthMiddleware)
	auth.Get("/files/list", handler.ListFiles)
	auth.Post("/files/upload", handler.UploadFile)
	auth.Get("/files/download", handler.DownloadFile)
	auth.Delete("/files/delete", handler.DeleteFile)
	auth.Put("/files/rename", handler.RenameFile)
	auth.Post("/folders/create", handler.CreateFolder)
	auth.Get("/media/public/share", handler.GeneratePublicLink)
	auth.Get("/media/stream", handler.StreamVideo)
}

func registerPublicRoutes(app *fiber.App) {
	// публичный доступ — без авторизации
	public := app.Group("/public")
	public.Get("/media/stream", handler.StreamPublicVideo)
}
