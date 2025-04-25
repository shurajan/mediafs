package service

import (
	"fmt"
	"github.com/grafov/m3u8"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CutService struct {
	BaseDir string // e.g., "videos"
}

func NewCutService(baseDir string) *CutService {
	return &CutService{BaseDir: baseDir}
}

func (s *CutService) CreateClip(videoname string, from, to int, name string) (string, error) {
	dir := filepath.Join(s.BaseDir, videoname)
	srcM3U8 := filepath.Join(dir, "playlist.m3u8")

	if name == "" {
		name = fmt.Sprintf("cut_%d_%d_%s", from, to, time.Now().Format("150405"))
	}
	destM3U8 := filepath.Join(dir, name+".m3u8")

	data, err := os.ReadFile(srcM3U8)
	if err != nil {
		return "", fmt.Errorf("failed to read playlist: %w", err)
	}

	playlist, listType, err := m3u8.DecodeFrom(strings.NewReader(string(data)), true)
	if err != nil {
		return "", fmt.Errorf("failed to parse playlist: %w", err)
	}

	mediaPL, ok := playlist.(*m3u8.MediaPlaylist)
	if !ok || listType != m3u8.MEDIA {
		return "", fmt.Errorf("not a valid media playlist")
	}

	// Make sure range is valid
	if from < 0 || to > int(mediaPL.Count()) || from >= to {
		return "", fmt.Errorf("invalid segment range: from=%d to=%d", from, to)
	}

	newPL, err := m3u8.NewMediaPlaylist(uint(to-from), uint(to-from))
	if err != nil {
		return "", fmt.Errorf("failed to create new media playlist: %w", err)
	}
	newPL.SeqNo = uint64(from)

	for i := from; i < to; i++ {
		seg := mediaPL.Segments[i]
		if seg == nil {
			continue
		}
		_ = newPL.AppendSegment(seg)
	}

	newPL.Close()

	if err := os.WriteFile(destM3U8, []byte(newPL.Encode().String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write cut playlist: %w", err)
	}

	return name + ".m3u8", nil
}
