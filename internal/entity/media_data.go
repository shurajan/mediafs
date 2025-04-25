package entity

import (
	"fmt"
	"os"
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
	return fmt.Sprintf("/videos/%s/", m.Folder)
}

func (m *MediaInfo) Playlists() []*Playlist {
	entries, err := os.ReadDir(m.EntryPath)
	if err != nil {
		return nil
	}

	var master, playlist *Playlist
	var others []*Playlist

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".m3u8") {
			continue
		}
		full := filepath.Join(m.EntryPath, name)
		switch name {
		case "master.m3u8":
			master = &Playlist{Path: full}
		case "playlist.m3u8":
			playlist = &Playlist{Path: full}
		default:
			others = append(others, &Playlist{Path: full})
		}
	}

	var result []*Playlist
	if master != nil {
		result = append(result, master)
	}
	if playlist != nil {
		result = append(result, playlist)
	}
	result = append(result, others...)
	return result
}
