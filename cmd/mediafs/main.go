package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/grandcat/zeroconf"

	"mediafs/internal/handler"
	"mediafs/internal/middleware"
)

const (
	token = "supersecret"
	port  = ":8000"
)

func main() {
	baseDir := ensureMediaFS()

	app := fiber.New()
	app.Use(middleware.BearerAuth(token))

	app.Get("/files", handler.ListFiles(baseDir))
	app.Get("/files/:filename", handler.StreamFile(baseDir))
	app.Delete("/files/:filename", handler.DeleteFile(baseDir))

	go publishBonjour()

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

func publishBonjour() {
	server, err := zeroconf.Register(
		"MediaFS",
		"_http._tcp",
		"local.",
		8000,
		nil,
		nil,
	)
	if err != nil {
		log.Println("❌ Failed to publish Bonjour service:", err)
		return
	}
	log.Println("✅ Bonjour service 'MediaFS._http._tcp.local' published")

	<-make(chan struct{})
	defer server.Shutdown()
}
