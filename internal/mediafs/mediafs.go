package mediafs

import (
	"os"
	"path/filepath"
)

var (
	BaseDir        string
	ServiceFolders = []string{".cache", ".hls", ".meta"}
)

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	BaseDir = filepath.Join(home, ".mediafs")

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
