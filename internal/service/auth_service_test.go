package service

import (
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) *AuthService {
	t.Helper()
	return NewAuthService(filepath.Join(t.TempDir(), "auth.json"))
}

func TestSetAndCheckPassword(t *testing.T) {
	a := newTestService(t)
	if err := a.SetPassword("s3cret"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if !a.CheckPassword("s3cret") {
		t.Error("expected correct password to validate")
	}
	if a.CheckPassword("wrong") {
		t.Error("expected wrong password to be rejected")
	}
}

func TestSetPasswordPersists(t *testing.T) {
	a := newTestService(t)
	if err := a.SetPassword("hunter2"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	reloaded := NewAuthService(a.path)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reloaded.CheckPassword("hunter2") {
		t.Error("password was not persisted across reload")
	}
}

func TestLoadMissingFile(t *testing.T) {
	a := NewAuthService(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err := a.Load(); err == nil {
		t.Error("expected error loading a missing file")
	}
}

func TestGenerateAndCheckToken(t *testing.T) {
	a := newTestService(t)
	if err := a.SetPassword("pw"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	token := a.GenerateToken()
	if token == "" {
		t.Fatal("expected a non-empty token")
	}
	if !a.CheckToken(token) {
		t.Error("expected the generated token to validate")
	}
	if a.CheckToken("not-the-token") {
		t.Error("expected a wrong token to be rejected")
	}
}

func TestGenerateTokenPersists(t *testing.T) {
	a := newTestService(t)
	if err := a.SetPassword("pw"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	token := a.GenerateToken()

	reloaded := NewAuthService(a.path)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reloaded.CheckToken(token) {
		t.Error("token was not persisted across reload")
	}
}

func TestSetPasswordResetsToken(t *testing.T) {
	a := newTestService(t)
	if err := a.SetPassword("pw"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	token := a.GenerateToken()

	// Changing the password must invalidate the previously issued token.
	if err := a.SetPassword("new-pw"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if a.CheckToken(token) {
		t.Error("expected old token to be invalidated after password change")
	}
}
