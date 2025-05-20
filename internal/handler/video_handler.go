package handler

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/entity"
	"os"
	"path/filepath"
	"strings"
)

type MediaFile struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	HLSURL             string  `json:"hlsURL"`
	KeyframesURL       *string `json:"keyframesURL,omitempty"`
	NsfwframesURL      *string `json:"nsfwframesURL,omitempty"`
	CreatedAt          string  `json:"createdAt,omitempty"`
	Duration           int     `json:"duration"`
	Resolution         string  `json:"resolution,omitempty"`
	SizeMB             int     `json:"sizeMB,omitempty"`
	SegmentCount       int     `json:"segmentCount"`
	AvgSegmentDuration float64 `json:"avgSegmentDuration"`
}

func ListVideos(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		files := make([]MediaFile, 0)

		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			folderName := entry.Name()
			info := entity.NewMediaInfo(baseDir, folderName)

			playlist := info.Playlist()
			if playlist == nil {
				continue
			}

			files = append(files, MediaFile{
				ID:                 info.ID(),
				Name:               folderName,
				HLSURL:             info.StreamURL(),
				KeyframesURL:       info.KeyFramesURL(),
				NsfwframesURL:      info.NsfwFramesURL(),
				CreatedAt:          info.CreatedAt(),
				Duration:           playlist.Duration(),
				Resolution:         playlist.Resolution(),
				SizeMB:             playlist.SizeMB(),
				SegmentCount:       playlist.SegmentCount(),
				AvgSegmentDuration: playlist.AvgSegmentDuration(),
			})
		}

		return c.JSON(files)
	}
}

// StreamHLSFile - теперь умеет правильно ставить Content-Type для mp4, jpg, vtt
func StreamHLSFile(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		relativePath := filepath.Clean(c.Params("*")) // <-- относительный путь внутри видео папки

		fullDir := filepath.Join(baseDir, videoname)
		fullPath := filepath.Join(fullDir, relativePath)

		if !strings.HasPrefix(fullPath, fullDir) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "invalid path",
			})
		}

		if _, err := os.Stat(fullPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "file not found",
			})
		}

		ext := strings.ToLower(filepath.Ext(fullPath))
		switch ext {
		case ".m3u8":
			c.Response().Header.Set("Content-Type", "application/vnd.apple.mpegurl")
		case ".ts":
			c.Response().Header.Set("Content-Type", "video/MP2T")
		case ".jpg", ".jpeg":
			c.Response().Header.Set("Content-Type", "image/jpeg")
		case ".mp4":
			c.Response().Header.Set("Content-Type", "video/mp4")
		case ".vtt":
			c.Response().Header.Set("Content-Type", "text/vtt")
		case ".zip":
			c.Response().Header.Set("Content-Type", "application/zip")
		default:
			c.Response().Header.Set("Content-Type", "application/octet-stream")
		}

		return c.SendFile(fullPath)
	}
}

func DeleteVideo(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		fullPath := filepath.Join(baseDir, videoname)

		if err := os.RemoveAll(fullPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to delete",
			})
		}

		return c.JSON(fiber.Map{
			"message": "deleted",
		})
	}
}
