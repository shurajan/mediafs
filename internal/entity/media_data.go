package entity

import (
	"fmt"
	"os"
	"path/filepath"
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

func (m *MediaInfo) ID() string {
	return simpleID(m.EntryPath)
}

func (m *MediaInfo) CreatedAt() string {
	info, err := os.Stat(m.EntryPath)
	if err != nil {
		return ""
	}
	return info.ModTime().UTC().Format(time.RFC3339)
}

func (m *MediaInfo) StreamURL() string {
	return fmt.Sprintf("/videos/%s/playlist.m3u8", m.Folder)
}

func (m *MediaInfo) KeyframesURL() *string {
	keyframesPath := filepath.Join(m.EntryPath, "keyframes")
	info, err := os.Stat(keyframesPath)
	if err != nil || !info.IsDir() {
		return nil
	}
	url := fmt.Sprintf("/keyframe/%s/", m.Folder)
	return &url
}

func (m *MediaInfo) Playlist() *Playlist {
	playlistPath := filepath.Join(m.EntryPath, "playlist.m3u8")
	if _, err := os.Stat(playlistPath); err == nil {
		return &Playlist{Path: playlistPath}
	}
	return nil
}
