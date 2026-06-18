package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

// generousTest gives handlers that may shell out to ffprobe time to finish.
var generousTest = fiber.TestConfig{Timeout: 30 * time.Second, FailOnTimeout: true}

// makeVideoDir creates baseDir/<name> containing a minimal HLS playlist.
func makeVideoDir(t *testing.T, baseDir, name string) {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	playlist := "#EXTM3U\n" +
		"#EXT-X-VERSION:3\n" +
		"#EXT-X-TARGETDURATION:10\n" +
		"#EXT-X-MEDIA-SEQUENCE:0\n" +
		"#EXTINF:9.000,\n" +
		"segment0.ts\n" +
		"#EXT-X-ENDLIST\n"
	if err := os.WriteFile(filepath.Join(dir, "playlist.m3u8"), []byte(playlist), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "segment0.ts"), []byte("ts-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListVideos(t *testing.T) {
	base := t.TempDir()
	makeVideoDir(t, base, "alpha")
	makeVideoDir(t, base, "beta")
	// A hidden dir and a plain file must be ignored.
	if err := os.MkdirAll(filepath.Join(base, ".hidden"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "loose.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A directory without a playlist must be skipped.
	if err := os.MkdirAll(filepath.Join(base, "no-playlist"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Get("/videos", ListVideos(base))

	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	resp, err := app.Test(req, generousTest)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var files []MediaFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d videos, want 2 (%+v)", len(files), files)
	}
	for _, f := range files {
		if f.Name != "alpha" && f.Name != "beta" {
			t.Errorf("unexpected video %q", f.Name)
		}
		if f.HLSURL == "" || f.ID == "" {
			t.Errorf("video %q is missing id/url: %+v", f.Name, f)
		}
	}
}

func TestStreamHLSFilePlaylist(t *testing.T) {
	base := t.TempDir()
	makeVideoDir(t, base, "alpha")

	app := fiber.New()
	app.Get("/videos/:videoname/*", StreamHLSFile(base))

	req := httptest.NewRequest(http.MethodGet, "/videos/alpha/playlist.m3u8", nil)
	resp, err := app.Test(req, generousTest)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/vnd.apple.mpegurl" {
		t.Errorf("Content-Type = %q, want application/vnd.apple.mpegurl", ct)
	}
}

func TestStreamHLSFileSegmentContentType(t *testing.T) {
	base := t.TempDir()
	makeVideoDir(t, base, "alpha")

	app := fiber.New()
	app.Get("/videos/:videoname/*", StreamHLSFile(base))

	req := httptest.NewRequest(http.MethodGet, "/videos/alpha/segment0.ts", nil)
	resp, err := app.Test(req, generousTest)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "video/MP2T" {
		t.Errorf("Content-Type = %q, want video/MP2T", ct)
	}
}

func TestStreamHLSFileNotFound(t *testing.T) {
	base := t.TempDir()
	makeVideoDir(t, base, "alpha")

	app := fiber.New()
	app.Get("/videos/:videoname/*", StreamHLSFile(base))

	req := httptest.NewRequest(http.MethodGet, "/videos/alpha/missing.m3u8", nil)
	resp, err := app.Test(req, generousTest)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}
}

func TestDeleteVideo(t *testing.T) {
	base := t.TempDir()
	makeVideoDir(t, base, "alpha")

	app := fiber.New()
	app.Delete("/videos/:videoname", DeleteVideo(base))

	req := httptest.NewRequest(http.MethodDelete, "/videos/alpha", nil)
	resp, err := app.Test(req, generousTest)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if _, err := os.Stat(filepath.Join(base, "alpha")); !os.IsNotExist(err) {
		t.Errorf("expected video dir to be removed, stat err = %v", err)
	}
}
