package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"path/filepath"

	"mediafs/internal/handler"
	"mediafs/internal/middleware"
	"mediafs/internal/service"
)

const port = ":8000"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash-password" {
		handlePasswordHashing()
		return
	}

	baseDir := ensureMediaFS()
	authPath := filepath.Join(baseDir, "auth.json")
	authService := service.NewAuthService(authPath)

	if err := authService.Load(); err != nil {
		log.Fatal("❌ auth.json not found. Create it first with `mediafs hash-password`.")
	}

	app := fiber.New()

	app.Post("/auth", handler.AuthHandler(authService))
	app.Use(middleware.BearerAuthMiddleware(authService))

	app.Get("/files", handler.ListFiles(baseDir))
	app.Get("/files/:filename", handler.StreamFile(baseDir))
	app.Delete("/files/:filename", handler.DeleteFile(baseDir))

	go service.PublishBonjour()

	log.Println("📡 MediaFS running on " + port)
	log.Fatal(app.Listen(port))
}

func ensureMediaFS() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Can't find home dir:", err)
	}
	path := filepath.Join(home, ".mediafs")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatal("Can't create mediafs dir:", err)
	}
	return path
}

func handlePasswordHashing() {
	hashCmd := flag.NewFlagSet("hash-password", flag.ExitOnError)
	passwordPtr := hashCmd.String("password", "", "Password to hash and save")
	_ = hashCmd.Parse(os.Args[2:])
	if *passwordPtr == "" {
		log.Fatal("❌ Please provide --password")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Can't find home dir:", err)
	}
	authPath := filepath.Join(home, ".mediafs", "auth.json")

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
