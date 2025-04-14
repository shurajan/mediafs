package service

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mediafs/internal/mediafs"
	"mediafs/internal/util"
)

type FileEntry struct {
	Name     string    `json:"name"`
	IsDir    bool      `json:"is_dir"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	FullPath string    `json:"path"`
}

func ListFiles(relPath string) ([]FileEntry, error) {
	dirPath, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		log.Printf("[ListFiles] invalid path %q: %v", relPath, err)
		return nil, err
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("[ListFiles] failed to read dir %q: %v", dirPath, err)
		return nil, err
	}

	var files []FileEntry
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Printf("[ListFiles] failed to get info for %q: %v", name, err)
			continue
		}
		files = append(files, FileEntry{
			Name:     name,
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			FullPath: filepath.Join(relPath, name),
		})
	}
	return files, nil
}

func UploadFile(relPath string, data io.Reader) error {
	path, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		log.Printf("[UploadFile] invalid path %q: %v", relPath, err)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Printf("[UploadFile] mkdir failed for %q: %v", path, err)
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("[UploadFile] create failed for %q: %v", path, err)
		return err
	}
	defer f.Close()
	if _, err = io.Copy(f, data); err != nil {
		log.Printf("[UploadFile] write failed for %q: %v", path, err)
	}
	return err
}

func DeleteFile(relPath string) error {
	path, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		log.Printf("[DeleteFile] invalid path %q: %v", relPath, err)
		return err
	}
	if err := os.Remove(path); err != nil {
		log.Printf("[DeleteFile] remove failed for %q: %v", path, err)
		return err
	}
	return nil
}

func RenameFile(oldRel, newRel string) error {
	oldPath, err := util.ResolveSafePath(mediafs.BaseDir, oldRel)
	if err != nil {
		log.Printf("[RenameFile] invalid old path %q: %v", oldRel, err)
		return err
	}
	newPath, err := util.ResolveSafePath(mediafs.BaseDir, newRel)
	if err != nil {
		log.Printf("[RenameFile] invalid new path %q: %v", newRel, err)
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		log.Printf("[RenameFile] rename failed %q → %q: %v", oldPath, newPath, err)
		return err
	}
	return nil
}

func ReadFile(relPath string) (*os.File, error) {
	path, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		log.Printf("[ReadFile] invalid path %q: %v", relPath, err)
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[ReadFile] open failed for %q: %v", path, err)
	}
	return f, err
}

func CreateFolder(relPath string) error {
	path, err := util.ResolveSafePath(mediafs.BaseDir, relPath)
	if err != nil {
		log.Printf("[CreateFolder] invalid path %q: %v", relPath, err)
		return err
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("[CreateFolder] mkdir failed for %q: %v", path, err)
		return err
	}
	return nil
}
