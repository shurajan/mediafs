package entity

import (
	"github.com/grafov/m3u8"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Playlist struct {
	Path   string
	cached *cachedInfo
}

type cachedInfo struct {
	duration   int
	sizeMB     int
	resolution string
	loaded     bool
}

func (p *Playlist) ID() string {
	data, err := os.ReadFile(p.Path)
	if err != nil {
		return simpleID(p.Path)
	}
	combined := p.Path + string(data)
	return simpleID(combined)
}

func (p *Playlist) Name() string {
	return filepath.Base(p.Path)
}

func (p *Playlist) Duration() int {
	p.ensureCached()
	return p.cached.duration
}

func (p *Playlist) SizeMB() int {
	p.ensureCached()
	return p.cached.sizeMB
}

func (p *Playlist) Resolution() string {
	p.ensureCached()
	return p.cached.resolution
}

func (p *Playlist) ensureCached() {
	if p.cached != nil && p.cached.loaded {
		return
	}

	duration := 0
	size := int64(0)
	var resolution string

	// Parse media playlist
	pl, err := p.parseMediaPlaylist()
	if err == nil {
		dir := filepath.Dir(p.Path)
		for _, seg := range pl.Segments {
			if seg == nil || seg.URI == "" {
				continue
			}
			duration += int(seg.Duration)

			tsPath := filepath.Join(dir, seg.URI)
			if info, err := os.Stat(tsPath); err == nil && !info.IsDir() {
				size += info.Size()
			}
		}
	}

	resolution = p.extractResolutionFromPlaylist()
	if resolution == "" {
		resolution = p.FFProbeResolution()
	}

	p.cached = &cachedInfo{
		duration:   duration,
		sizeMB:     int(math.Round(float64(size) / 1024.0 / 1024.0)),
		resolution: resolution,
		loaded:     true,
	}
}

func (p *Playlist) extractResolutionFromPlaylist() string {
	f, err := os.Open(p.Path)
	if err != nil {
		return ""
	}
	defer f.Close()

	master, listType, err := m3u8.DecodeFrom(f, true)
	if err != nil || listType != m3u8.MASTER {
		return ""
	}

	if mpl, ok := master.(*m3u8.MasterPlaylist); ok {
		for _, variant := range mpl.Variants {
			if variant.Resolution != "" {
				return variant.Resolution
			}
		}
	}
	return ""
}

func (p *Playlist) FFProbeResolution() string {
	pl, err := p.parseMediaPlaylist()
	if err != nil {
		return ""
	}

	dir := filepath.Dir(p.Path)
	for _, seg := range pl.Segments {
		if seg == nil || seg.URI == "" {
			continue
		}

		tsPath := filepath.Join(dir, seg.URI)
		if _, err := os.Stat(tsPath); err == nil {
			cmd := exec.Command("ffprobe",
				"-v", "error",
				"-select_streams", "v:0",
				"-show_entries", "stream=width,height",
				"-of", "csv=p=0:s=x",
				tsPath,
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
		}
	}
	return ""
}

func (p *Playlist) parseMediaPlaylist() (*m3u8.MediaPlaylist, error) {
	count := estimateSegmentCount(p.Path)

	f, err := os.Open(p.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pl, err := m3u8.NewMediaPlaylist(uint(count), uint(count))
	if err != nil {
		return nil, err
	}
	err = pl.DecodeFrom(f, true)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func estimateSegmentCount(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 100
	}
	count := 0
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "#EXTINF:") {
			count++
		}
	}
	if count < 10 {
		return 50
	}
	return count + 20
}
