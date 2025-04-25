package handler

import (
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/entity"
	"os"
	"path/filepath"
	"strings"
)

type PlaylistInfo struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Duration           int     `json:"duration"`
	Resolution         string  `json:"resolution,omitempty"`
	SizeMB             int     `json:"sizeMB,omitempty"`
	SegmentCount       int     `json:"segmentCount"`
	AvgSegmentDuration float64 `json:"avgSegmentDuration"`
}

type MediaFile struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	HLSURL    string         `json:"hlsURL"`
	CreatedAt string         `json:"createdAt,omitempty"`
	Playlists []PlaylistInfo `json:"playlists"`
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

			playlists := info.Playlists()
			if len(playlists) == 0 {
				continue
			}

			var playlistInfos []PlaylistInfo
			for _, p := range playlists {
				playlistInfos = append(playlistInfos, PlaylistInfo{
					ID:                 p.ID(),
					Name:               p.Name(),
					Duration:           p.Duration(),
					Resolution:         p.Resolution(),
					SizeMB:             p.SizeMB(),
					SegmentCount:       p.SegmentCount(),
					AvgSegmentDuration: p.AvgSegmentDuration(),
				})
			}

			files = append(files, MediaFile{
				ID:        info.ID(),
				Name:      folderName,
				HLSURL:    info.StreamURL(),
				CreatedAt: info.CreatedAt(),
				Playlists: playlistInfos,
			})
		}

		return c.JSON(files)
	}
}

func StreamHLSPlaylist(baseDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		playlist := filepath.Base(c.Params("playlist"))
		if playlist == "" {
			playlist = "playlist.m3u8"
		}

		playlistPath := filepath.Join(baseDir, videoname, playlist)
		if _, err := os.Stat(playlistPath); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "playlist not found",
			})
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "segment not found",
			})
		}

		return c.SendFile(segmentPath)
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
