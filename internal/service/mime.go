package service

import (
	"mime"
	"path/filepath"
	"strings"
)

func DetectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType
	}

	// fallback
	switch ext {
	case ".mkv":
		return "video/x-matroska"
	case ".ts":
		return "video/mp2t"
	case ".mov":
		return "video/quicktime"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".flv":
		return "video/x-flv"
	}

	return "application/octet-stream"
}
