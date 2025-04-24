package handler

import (
	"encoding/base32"
	"fmt"
	"github.com/zeebo/blake3"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type MediaInfo struct {
	BaseDir   string
	Folder    string
	EntryPath string
}

func NewMediaInfo(baseDir, folder string) *MediaInfo {
	return &MediaInfo{
		BaseDir:   baseDir,
		Folder:    folder,
		EntryPath: filepath.Join(baseDir, folder),
	}
}

// StreamURL returns the preferred playlist URL to serve (master if available)
func (m *MediaInfo) StreamURL() string {
	if m.hasMaster() {
		return fmt.Sprintf("/videos/%s/master.m3u8", m.Folder)
	}
	return fmt.Sprintf("/videos/%s/playlist.m3u8", m.Folder)
}

// MasterPath returns path to master.m3u8 if it exists
func (m *MediaInfo) MasterPath() string {
	return filepath.Join(m.EntryPath, "master.m3u8")
}

// PlaylistPath returns playlist path to use for metadata (master if available)
func (m *MediaInfo) PlaylistPath() string {
	if m.hasMaster() {
		return m.MasterPath()
	}
	return filepath.Join(m.EntryPath, "playlist.m3u8")
}

func (m *MediaInfo) hasMaster() bool {
	_, err := os.Stat(m.MasterPath())
	return err == nil
}

func (m *MediaInfo) ID() string {
	mediaPath := filepath.Join(m.EntryPath, "playlist.m3u8")

	data, err := os.ReadFile(mediaPath)
	if err != nil {
		return m.simpleID()
	}

	source := []byte(m.EntryPath)
	source = append(source, data...)

	hash := blake3.Sum256(source)
	sum := hash[:16]

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)
}

func (m *MediaInfo) simpleID() string {
	name := strings.ToLower(strings.TrimSpace(m.Folder))
	hash := blake3.Sum256([]byte(name))
	sum := hash[:8]
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)
}

func (m *MediaInfo) Duration() int {
	content, err := os.ReadFile(filepath.Join(m.EntryPath, "playlist.m3u8")) // важно: для длительности используем media playlist
	if err != nil {
		return 0
	}
	lines := strings.Split(string(content), "\n")
	total := 0.0
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXTINF:") {
			var sec float64
			fmt.Sscanf(line, "#EXTINF:%f,", &sec)
			total += sec
		}
	}
	return int(total)
}

func (m *MediaInfo) Resolution() string {
	if res := m.resolutionFromPlaylist(); res != "" {
		return res
	}
	return m.resolutionFromFFProbe()
}

func (m *MediaInfo) resolutionFromPlaylist() string {
	data, err := os.ReadFile(m.PlaylistPath())
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") && strings.Contains(line, "RESOLUTION=") {
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "RESOLUTION=") {
					return strings.TrimPrefix(part, "RESOLUTION=")
				}
			}
		}
	}
	return ""
}

func (m *MediaInfo) resolutionFromFFProbe() string {
	segments, err := filepath.Glob(filepath.Join(m.EntryPath, "*.ts"))
	if err != nil || len(segments) == 0 {
		return ""
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		segments[0],
	)

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func (m *MediaInfo) CreatedAt() string {
	info, err := os.Stat(m.EntryPath)
	if err != nil {
		return ""
	}
	return info.ModTime().UTC().Format(time.RFC3339)
}

func (m *MediaInfo) SizeMB() int {
	var total int64
	_ = filepath.Walk(m.EntryPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".ts") {
			total += info.Size()
		}
		return nil
	})
	return int(math.Round(float64(total) / 1024.0 / 1024.0))
}
