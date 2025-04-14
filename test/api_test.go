package test

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"mediafs/internal/handler"
	"mediafs/internal/middleware"
	"mediafs/internal/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupApp() *fiber.App {
	_ = os.Setenv("MEDIAFS_MODE", "test")
	_ = service.InitMediaFS()
	_ = service.InitTestEnv()

	app := fiber.New()
	app.Use(middleware.AuthMiddleware)

	app.Get("/files/list", handler.ListFiles)
	app.Post("/files/upload", handler.UploadFile)
	app.Get("/files/download", handler.DownloadFile)
	app.Delete("/files/delete", handler.DeleteFile)
	app.Put("/files/rename", handler.RenameFile)
	app.Post("/folders/create", handler.CreateFolder)
	return app
}

func do(app *fiber.App, method, url string, body io.Reader, contentType string) *http.Response {
	req := httptest.NewRequest(method, url, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Authorization", "Bearer test-token-123")
	resp, _ := app.Test(req)
	return resp
}

func TestCreateListFolder(t *testing.T) {
	app := setupApp()
	path := "newfolder"
	_ = os.RemoveAll(filepath.Join(service.TestFilesDir, path))

	body := strings.NewReader(`{"path": "` + path + `"}`)
	resp := do(app, "POST", "/folders/create", body, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create failed: %d", resp.StatusCode)
	}

	resp = do(app, "GET", "/files/list?path=newfolder", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list failed: %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if string(bodyBytes) != "[]" {
		t.Errorf("expected [], got %s", string(bodyBytes))
	}
}

func TestUploadDownloadDelete(t *testing.T) {
	app := setupApp()

	srcPath := filepath.Join(service.TestFilesDir, "texts", "hello.txt")
	dstRelPath := "testdata/upload.txt"
	dstAbsPath := filepath.Join(service.TestFilesDir, dstRelPath)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	srcFile, _ := os.Open(srcPath)
	w, _ := writer.CreateFormFile("file", "hello.txt")
	io.Copy(w, srcFile)
	srcFile.Close()
	writer.Close()

	resp := do(app, "POST", "/files/upload?path="+dstRelPath, &buf, writer.FormDataContentType())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload failed: %d", resp.StatusCode)
	}

	resp = do(app, "GET", "/files/download?path="+dstRelPath, nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download failed: %d", resp.StatusCode)
	}

	resp = do(app, "DELETE", "/files/delete?path="+dstRelPath, nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete failed: %d", resp.StatusCode)
	}

	if _, err := os.Stat(dstAbsPath); err == nil {
		t.Error("file was not deleted")
	}
}

func TestRename(t *testing.T) {
	app := setupApp()
	src := "testdata/old.txt"
	dst := "testdata/new.txt"

	// гарантируем наличие исходного файла
	srcAbs := filepath.Join(service.TestFilesDir, src)
	_ = os.MkdirAll(filepath.Dir(srcAbs), 0755)
	_ = os.WriteFile(srcAbs, []byte("rename test"), 0644)

	payload := map[string]string{"old_path": src, "new_path": dst}
	jsonBody, _ := json.Marshal(payload)

	resp := do(app, "PUT", "/files/rename", bytes.NewReader(jsonBody), "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("rename failed: %d", resp.StatusCode)
	}

	dstAbs := filepath.Join(service.TestFilesDir, dst)
	if _, err := os.Stat(dstAbs); err != nil {
		t.Fatal("renamed file not found")
	}
}

func TestErrorCases(t *testing.T) {
	app := setupApp()

	resp := do(app, "GET", "/files/download?path=notfound.txt", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for missing file, got %d", resp.StatusCode)
	}

	resp = do(app, "DELETE", "/files/delete?path=notfound.txt", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 500 for delete missing, got %d", resp.StatusCode)
	}

	resp = do(app, "POST", "/files/upload?path=err.txt", nil, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing file, got %d", resp.StatusCode)
	}

	resp = do(app, "PUT", "/files/rename", strings.NewReader("invalid"), "application/json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", resp.StatusCode)
	}

	resp = do(app, "GET", "/files/list?path=../../", nil, "")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected error for outside base dir, got %d", resp.StatusCode)
	}

	resp = do(app, "DELETE", "/files/delete?path=/etc/passwd", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected error for forbidden absolute path, got %d", resp.StatusCode)
	}
}

func TestAuth(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/files/list?path=texts", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing Authorization header, got %d", resp.StatusCode)
	}

	req = httptest.NewRequest("GET", "/files/list?path=texts", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token, got %d", resp.StatusCode)
	}

	req = httptest.NewRequest("GET", "/files/list?path=texts", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for valid token, got %d", resp.StatusCode)
	}
}
