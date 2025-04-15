package test

import (
	"bytes"
	"io"
	"mediafs/internal/handler"
	"mediafs/internal/middleware"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
)

const (
	testToken  = "supersecret"
	testDir    = ".mediafs_test"
	testFile   = "testfile.mp4"
	testFileTS = "teststream.ts"
)

func setupTestApp(baseDir string) *fiber.App {
	app := fiber.New()
	app.Use(middleware.BearerAuth(testToken))
	app.Get("/files", handler.ListFiles(baseDir))
	app.Get("/files/:filename", handler.StreamFile(baseDir))
	app.Delete("/files/:filename", handler.DeleteFile(baseDir))
	return app
}

func prepareTestFiles(t *testing.T, baseDir string) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal("failed to create test dir:", err)
	}

	files := []string{testFile, testFileTS}
	for _, name := range files {
		path := filepath.Join(baseDir, name)
		err := os.WriteFile(path, []byte("dummy content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}
}

func cleanupTestDir(t *testing.T, baseDir string) {
	if err := os.RemoveAll(baseDir); err != nil {
		t.Log("failed to clean up:", err)
	}
}

func TestMediaFS(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal("Cannot determine home dir:", err)
	}
	baseDir := filepath.Join(home, testDir)

	prepareTestFiles(t, baseDir)
	defer cleanupTestDir(t, baseDir)

	app := setupTestApp(baseDir)

	// 1. Список файлов
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatal("List files failed:", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte(testFile)) {
		t.Error("test file not found in response")
	}

	// 2. Получение файла
	req = httptest.NewRequest(http.MethodGet, "/files/"+testFile, nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatal("Stream file failed:", err)
	}
	content, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(content, []byte("dummy content")) {
		t.Error("file content mismatch")
	}

	// 3. Удаление файла
	req = httptest.NewRequest(http.MethodDelete, "/files/"+testFileTS, nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatal("Delete file failed:", err)
	}

	if _, err := os.Stat(filepath.Join(baseDir, testFileTS)); !os.IsNotExist(err) {
		t.Error("file was not deleted")
	}
}
