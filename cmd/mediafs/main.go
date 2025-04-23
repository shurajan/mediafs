package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"mediafs/internal/handler"
	"mediafs/internal/middleware"
	"mediafs/internal/service"
)

const (
	port          = ":8000"
	cmdHashPasswd = "hash-password"
)

func main() {
	baseDir, metaDir := ensureMediaFS()

	if len(os.Args) > 1 && os.Args[1] == cmdHashPasswd {
		handlePasswordHashing(metaDir)
		return
	}
	authService := setupAuth(metaDir)
	// Настройка контекста для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация компонентов
	app := setupFiberApp(baseDir, authService)

	// WaitGroup для всех горутин
	var wg sync.WaitGroup

	// Запуск HTTP‑сервера
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("📡 MediaFS running on " + port)
		if err := app.Listen(port); err != nil {
			log.Printf("❌ Server error: %v", err)
			cancel()
		}
	}()

	// Обработка сигналов завершения
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала или отмены контекста
	select {
	case sig := <-sigs:
		log.Printf("🛑 Received signal: %v, shutting down...", sig)
	case <-ctx.Done():
		log.Println("🛑 Context canceled, shutting down...")
	}

	// Останавливаем HTTP‑сервер
	if err := app.Shutdown(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
	}

	// Ждем завершения всех горутин
	wg.Wait()
	log.Println("👋 Shutdown complete.")
}

// setupAuth настраивает сервис аутентификации
func setupAuth(metaDir string) *service.AuthService {
	authPath := filepath.Join(metaDir, "auth.json")
	authService := service.NewAuthService(authPath)

	if err := authService.Load(); err != nil {
		log.Fatal("❌ auth.json not found. Create it first with `mediafs hash-password`.")
	}

	return authService
}

// setupFiberApp настраивает Fiber‑приложение
func setupFiberApp(baseDir string, authService *service.AuthService) *fiber.App {
	app := fiber.New()

	// Аутентификация
	app.Post("/auth", handler.AuthHandler(authService))

	// Middleware авторизации
	app.Use(middleware.BearerAuthMiddleware(authService))

	// HLS-файловый сервис
	app.Get("/videos", handler.ListFiles(baseDir))                                 // список видео
	app.Get("/videos/:filename/playlist.m3u8", handler.StreamHLSPlaylist(baseDir)) // .m3u8
	app.Get("/videos/:filename/:segment", handler.StreamHLSSegment(baseDir))       // .ts
	app.Delete("/videos/:filename", handler.DeleteFile(baseDir))                   // удаление всей папки

	return app
}

// ensureMediaFS проверяет и создаёт директории ~/.mediafs и ~/.mediafs/.meta
func ensureMediaFS() (mediafsPath string, metaPath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Can't find home dir:", err)
	}

	mediafsPath = filepath.Join(home, ".mediafs")
	if err = os.MkdirAll(mediafsPath, 0o755); err != nil {
		log.Fatal("Can't create mediafs dir:", err)
	}

	metaPath = filepath.Join(mediafsPath, ".meta")
	if err = os.MkdirAll(metaPath, 0o755); err != nil {
		log.Fatal("Can't create .meta dir:", err)
	}

	return mediafsPath, metaPath
}

// handlePasswordHashing обрабатывает команду hash-password и получает путь к мета‑директории извне
func handlePasswordHashing(metaDir string) {
	hashCmd := flag.NewFlagSet(cmdHashPasswd, flag.ExitOnError)
	passwordPtr := hashCmd.String("password", "", "Password to hash and save")
	_ = hashCmd.Parse(os.Args[2:])

	if *passwordPtr == "" {
		log.Fatal("❌ Please provide --password")
	}

	authPath := filepath.Join(metaDir, "auth.json")
	authService := service.NewAuthService(authPath)

	if _, err := os.Stat(authPath); err == nil {
		var response string
		fmt.Printf("⚠️  %s already exists. Overwrite? Type 'yes' to confirm: ", authPath)
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("❌ Aborted.")
			return
		}
	}

	if err := authService.SetPassword(*passwordPtr); err != nil {
		log.Fatal("Failed to save password:", err)
	}
	fmt.Println("✅ Password hash saved to:", authPath)
}
