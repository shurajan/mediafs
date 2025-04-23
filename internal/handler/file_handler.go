package handler

import (
	"encoding/base32"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeebo/blake3"
)

func ListFiles(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		type MediaFile struct {
			ID     string `json:"id"`
			Name   string `json:"name"`   // Название папки
			HLSURL string `json:"hlsURL"` // Ссылка на .m3u8
		}

		files := make([]MediaFile, 0)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()

			// Исключаем системные и скрытые директории (начинаются с точки)
			if strings.HasPrefix(name, ".") {
				continue
			}

			playlistPath := filepath.Join(baseDir, name, "playlist.m3u8")
			if _, err := os.Stat(playlistPath); err != nil {
				continue
			}

			files = append(files, MediaFile{
				ID:     IDFromFolder(baseDir, name),
				Name:   name,
				HLSURL: fmt.Sprintf("/videos/%s/playlist.m3u8", name),
			})
		}

		return c.JSON(files)
	}
}

func StreamHLSPlaylist(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename"))
		playlistPath := filepath.Join(baseDir, filename, "playlist.m3u8")

		if _, err := os.Stat(playlistPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "playlist not found"})
		}
		return c.SendFile(playlistPath)
	}
}

func StreamHLSSegment(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename"))
		segment := filepath.Base(c.Params("segment"))
		segmentPath := filepath.Join(baseDir, filename, segment)

		if _, err := os.Stat(segmentPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "segment not found"})
		}
		return c.SendFile(segmentPath)
	}
}

func DeleteFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename"))
		fullPath := filepath.Join(baseDir, filename)

		if err := os.RemoveAll(fullPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete"})
		}

		return c.JSON(fiber.Map{"message": "deleted"})
	}
}

func IDFromFolder(baseDir, folderName string) string {
	playlistPath := filepath.Join(baseDir, folderName, "playlist.m3u8")

	data, err := os.ReadFile(playlistPath)
	if err != nil {
		return simpleID(folderName)
	}

	hash := blake3.Sum256(data)
	sum := hash[:16]
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString(sum)
}

func simpleID(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	hash := blake3.Sum256([]byte(name))
	sum := hash[:8]
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)
}
