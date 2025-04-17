package test

import (
	"bytes"
	"io"
	"mediafs/internal/handler"
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

	app.Use(func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth != "Bearer "+testToken {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	})

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

	t.Run("unauthorized access returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Error("expected 401 for missing token")
		}
	})

	t.Run("list files", func(t *testing.T) {
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
	})

	t.Run("stream file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/"+testFile, nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatal("Stream file failed:", err)
		}
		content, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(content, []byte("dummy content")) {
			t.Error("file content mismatch")
		}
	})

	t.Run("delete file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/files/"+testFileTS, nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatal("Delete file failed:", err)
		}

		if _, err := os.Stat(filepath.Join(baseDir, testFileTS)); !os.IsNotExist(err) {
			t.Error("file was not deleted")
		}
	})

	t.Run("stream deleted file returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/"+testFileTS, nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != http.StatusNotFound {
			t.Error("expected 404 after deleted file, got:", resp.StatusCode)
		}
	})

	t.Run("delete nonexistent file returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/files/missing.ts", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != http.StatusNotFound {
			t.Error("expected 404 for missing file")
		}
	})
}
