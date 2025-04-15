package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"mediafs/internal/service"
	"os"
	"path/filepath"
	"strings"
)

func ListFiles(c *fiber.Ctx) error {
	path := c.Query("path", ".")
	log.Println("📂 ListFiles called with path:", path)

	files, err := service.ListFiles(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("❌ Folder not found:", path)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "folder not found"})
		}
		log.Println("❌ Error listing files:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if files == nil {
		files = []service.FileEntry{}
	}

	return c.JSON(files)
}

func UploadFile(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}
	log.Println("📤 UploadFile to path:", path)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required"})
	}
	log.Println("├─ Uploaded filename:", file.Filename)

	opened, err := file.Open()
	if err != nil {
		log.Println("❌ Failed to open uploaded file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer opened.Close()

	if err := service.UploadFile(path, opened); err != nil {
		if os.IsNotExist(err) {
			log.Println("❌ Target folder not found:", path)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "target folder not found"})
		}
		log.Println("❌ Upload error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "uploaded"})
}

func DownloadFile(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}
	log.Println("📥 DownloadFile path:", path)

	f, err := service.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("❌ File not found:", path)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		log.Println("❌ ReadFile error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer f.Close()

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
	return c.SendFile(f.Name())
}

func DeleteFile(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path query required"})
	}
	log.Println("🗑 DeleteFile path:", path)

	if err := service.DeleteFile(path); err != nil {
		if os.IsNotExist(err) {
			log.Println("❌ File not found:", path)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		log.Println("❌ Delete error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}

func RenameFile(c *fiber.Ctx) error {
	type RenameRequest struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	var req RenameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	log.Println("✏️ RenameFile:", req.OldPath, "→", req.NewPath)

	if err := service.RenameFile(req.OldPath, req.NewPath); err != nil {
		if os.IsNotExist(err) {
			log.Println("❌ Rename failed, file not found:", req.OldPath)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
		}
		log.Println("❌ Rename error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "renamed"})
}

func CreateFolder(c *fiber.Ctx) error {
	type CreateRequest struct {
		Path string `json:"path"`
	}
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	log.Println("📁 CreateFolder:", req.Path)

	if err := service.CreateFolder(strings.TrimSpace(req.Path)); err != nil {
		log.Println("❌ CreateFolder error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "folder created"})
}
