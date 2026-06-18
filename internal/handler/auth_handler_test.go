package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"mediafs/internal/service"
)

func newAuthService(t *testing.T, password string) *service.AuthService {
	t.Helper()
	a := service.NewAuthService(filepath.Join(t.TempDir(), "auth.json"))
	if err := a.SetPassword(password); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	return a
}

func postAuth(t *testing.T, auth *service.AuthService, body string) *http.Response {
	t.Helper()
	app := fiber.New()
	app.Post("/auth", AuthHandler(auth))

	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

func TestAuthHandlerCorrectPassword(t *testing.T) {
	auth := newAuthService(t, "letmein")
	resp := postAuth(t, auth, `{"password":"letmein"}`)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Token == "" {
		t.Error("expected a token in the response")
	}
	if !auth.CheckToken(body.Token) {
		t.Error("returned token should validate against the service")
	}
}

func TestAuthHandlerWrongPassword(t *testing.T) {
	auth := newAuthService(t, "letmein")
	resp := postAuth(t, auth, `{"password":"nope"}`)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAuthHandlerInvalidJSON(t *testing.T) {
	auth := newAuthService(t, "letmein")
	resp := postAuth(t, auth, `{not-json`)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}
