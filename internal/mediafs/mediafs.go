package mediafs

import (
	"os"
	"path/filepath"
)

var (
	BaseDir        string
	ServiceFolders = []string{".cache", ".hls", ".meta"}
)

// Init инициализирует базовую директорию и служебные подпапки
func Init(base string) error {
	BaseDir = base
	if err := os.MkdirAll(BaseDir, 0755); err != nil {
		return err
	}
	for _, dir := range ServiceFolders {
		if err := os.MkdirAll(filepath.Join(BaseDir, dir), 0755); err != nil {
			return err
		}
	}
	return nil
}
