package service

import (
	"fmt"
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

	content, err := os.ReadFile(srcM3U8)
	if err != nil {
		return "", fmt.Errorf("failed to read playlist: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var segments []string
	segmentIndex := 0
	var maxDuration float64
	var mediaSequence int

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "#EXTINF") {
			if segmentIndex == from {
				mediaSequence = segmentIndex
			}
			if segmentIndex >= from && segmentIndex < to {
				segments = append(segments, line)

				// parse duration
				var duration float64
				fmt.Sscanf(line, "#EXTINF:%f,", &duration)
				if duration > maxDuration {
					maxDuration = duration
				}

				if i+1 < len(lines) {
					segments = append(segments, lines[i+1])
				}
				i++
			}
			segmentIndex++
		}
	}

	if len(segments) == 0 {
		return "", fmt.Errorf("no segments in range")
	}

	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	b.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(maxDuration)+1))
	b.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", mediaSequence))

	for _, line := range segments {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("#EXT-X-ENDLIST\n")

	if err := os.WriteFile(destM3U8, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write cut playlist: %w", err)
	}

	return name + ".m3u8", nil
}
