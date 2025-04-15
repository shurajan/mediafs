package test

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"mediafs/internal/handler"
	"mediafs/internal/mediafs"
	"mediafs/internal/middleware"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupChunkApp() *fiber.App {
	app := fiber.New()
	app.Post("/api/files/upload/chunk", middleware.AuthMiddleware, handler.ChunkUploadHandler)
	return app
}

func TestChunkUpload(t *testing.T) {
	mediafs.Init()
	app := setupChunkApp()

	relPath := "chunks/test_chunked.txt"
	absPath := filepath.Join(mediafs.BaseDir, relPath)
	chunk1 := []byte("Hello ")
	chunk2 := []byte("Chunk ")
	chunk3 := []byte("Upload!")

	chunks := [][]byte{chunk1, chunk2, chunk3}
	for i, chunk := range chunks {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		w, _ := writer.CreateFormFile("chunk", fmt.Sprintf("chunk%d", i))
		w.Write(chunk)
		writer.Close()

		isLast := "0"
		if i == len(chunks)-1 {
			isLast = "1"
		}

		url := fmt.Sprintf("/api/files/upload/chunk?path=%s&index=%d&last=%s", relPath, i, isLast)
		req := httptest.NewRequest("POST", url, &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token-123")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status on chunk %d: %d", i, resp.StatusCode)
		}
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("failed to read final file: %v", err)
	}
	expected := string(chunk1) + string(chunk2) + string(chunk3)
	if string(data) != expected {
		t.Errorf("upload result mismatch: got %q, want %q", string(data), expected)
	}

	// Cleanup
	_ = os.Remove(absPath)
	_ = os.Remove(absPath + ".upload")
}
