package util

import (
	"errors"
	"path/filepath"
	"strings"
)

// ResolveSafePath проверяет, что путь внутри baseDir
func ResolveSafePath(baseDir, relPath string) (string, error) {
	absPath := filepath.Clean(filepath.Join(baseDir, relPath))
	if !strings.HasPrefix(absPath, baseDir) {
		return "", errors.New("access outside of baseDir is forbidden")
	}
	return absPath, nil
}
