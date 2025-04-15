// internal/test/stream_test.go
package test

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mediafs/internal/handler"
	"mediafs/internal/mediafs"
	"mediafs/internal/middleware"
	"mediafs/internal/util"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupStreamApp() *fiber.App {
	app := fiber.New()

	api := app.Group("/api", middleware.AuthMiddleware)
	api.Get("/media/public/share", handler.GeneratePublicLink)
	api.Get("/media/stream", handler.StreamVideo)

	public := app.Group("/public")
	public.Get("/media/stream", handler.StreamPublicVideo)

	return app
}

func TestStreamPublicVideo(t *testing.T) {
	mediafs.Init()
	app := setupStreamApp()

	// Prepare video file
	relPath := "videos/sample.mp4"
	absPath := filepath.Join(mediafs.BaseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create video folder: %v", err)
	}
	content := []byte("test-video-content")
	if err := os.WriteFile(absPath, content, 0644); err != nil {
		t.Fatalf("failed to write video file: %v", err)
	}

	// Generate signed public link
	expires := time.Now().Add(1 * time.Hour).Unix()
	sig := util.GenerateSignature(relPath, expires)
	escaped := url.QueryEscape(relPath)
	fullURL := fmt.Sprintf("/public/media/stream?path=%s&expires=%d&sig=%s", escaped, expires, sig)

	req := httptest.NewRequest("GET", fullURL, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != string(content) {
		t.Errorf("response content mismatch: got %q, want %q", string(body), string(content))
	}
}

func TestGeneratePublicLink(t *testing.T) {
	mediafs.Init()
	app := setupStreamApp()

	// Prepare file
	relPath := "videos/test.mp4"
	absPath := filepath.Join(mediafs.BaseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(absPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/media/public/share?path="+url.QueryEscape(relPath), nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want 200", resp.StatusCode)
	}

	var data map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&data)
	if !strings.Contains(data["url"], "/public/media/stream") {
		t.Errorf("unexpected public stream URL: %s", data["url"])
	}
}
