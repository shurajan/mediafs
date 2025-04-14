package service

import (
	"io"
	"os"
)

func InitTestEnv() error {
	temporaryFiles := []string{
		"upload.txt",
		"err.txt",
		"testdata/new.txt",
	}
	temporaryDirs := []string{
		"newfolder",
	}

	for _, f := range temporaryFiles {
		p, _ := resolveSafePath(f)
		_ = os.Remove(p)
	}

	for _, d := range temporaryDirs {
		p, _ := resolveSafePath(d)
		_ = os.RemoveAll(p)
	}

	return os.WriteFile(TestTokenFile, []byte("test-token-123"), 0644)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
