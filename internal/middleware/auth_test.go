package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
	"mediafs/internal/service"
)

func newAuthWithToken(t *testing.T) (*service.AuthService, string) {
	t.Helper()
	a := service.NewAuthService(filepath.Join(t.TempDir(), "auth.json"))
	if err := a.SetPassword("pw"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	return a, a.GenerateToken()
}

func newProtectedApp(auth *service.AuthService) *fiber.App {
	app := fiber.New()
	app.Use(BearerAuthMiddleware(auth))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func doGet(t *testing.T, app *fiber.App, header string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

func TestBearerAuthValidToken(t *testing.T) {
	auth, token := newAuthWithToken(t)
	app := newProtectedApp(auth)

	resp := doGet(t, app, "Bearer "+token)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func TestBearerAuthMissingHeader(t *testing.T) {
	auth, _ := newAuthWithToken(t)
	app := newProtectedApp(auth)

	resp := doGet(t, app, "")
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestBearerAuthWrongToken(t *testing.T) {
	auth, _ := newAuthWithToken(t)
	app := newProtectedApp(auth)

	resp := doGet(t, app, "Bearer wrong-token")
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestBearerAuthMalformedHeader(t *testing.T) {
	auth, token := newAuthWithToken(t)
	app := newProtectedApp(auth)

	// Missing the "Bearer " prefix.
	resp := doGet(t, app, token)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}
