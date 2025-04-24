package handler

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"path/filepath"
	"strings"
)

type MediaFile struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	HLSURL     string   `json:"hlsURL"`
	Duration   int      `json:"duration,omitempty"`   // seconds
	Resolution string   `json:"resolution,omitempty"` // video quality
	CreatedAt  string   `json:"createdAt,omitempty"`  // RFC3339
	SizeMB     int      `json:"sizeMB,omitempty"`     // rounded megabytes
	Clips      []string `json:"clips,omitempty"`      // extra .m3u8 files (except playlist/master)
}

func ListVideos(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		files := make([]MediaFile, 0)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			name := entry.Name()
			info := NewMediaInfo(baseDir, name)

			if _, err := os.Stat(info.PlaylistPath()); err != nil {
				continue
			}

			files = append(files, MediaFile{
				ID:         info.ID(),
				Name:       name,
				HLSURL:     info.StreamURL(),
				Duration:   info.Duration(),
				Resolution: info.Resolution(),
				CreatedAt:  info.CreatedAt(),
				SizeMB:     info.SizeMB(),
				Clips:      info.ExtraPlaylists(),
			})
		}

		return c.JSON(files)
	}
}

func StreamHLSPlaylist(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		playlist := filepath.Base(c.Params("playlist")) // support e.g. cut_x_y.m3u8
		if playlist == "" {
			playlist = "playlist.m3u8"
		}

		playlistPath := filepath.Join(baseDir, videoname, playlist)
		if _, err := os.Stat(playlistPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "playlist not found"})
		}
		return c.SendFile(playlistPath)
	}
}

func StreamHLSSegment(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		segment := filepath.Base(c.Params("segment"))
		segmentPath := filepath.Join(baseDir, videoname, segment)

		if _, err := os.Stat(segmentPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "segment not found"})
		}
		return c.SendFile(segmentPath)
	}
}

func DeleteVideo(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		fullPath := filepath.Join(baseDir, videoname)

		if err := os.RemoveAll(fullPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete"})
		}

		return c.JSON(fiber.Map{"message": "deleted"})
	}
}
