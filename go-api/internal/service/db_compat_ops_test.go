package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBuildDefaultAdminPasswordHashUsesBcrypt(t *testing.T) {
	hashed, err := buildDefaultAdminPasswordHash()
	if err != nil {
		t.Fatalf("build hash: %v", err)
	}
	if hashed == "" || hashed == "admin123" {
		t.Fatalf("expected bcrypt hash, got %q", hashed)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte("admin123")); err != nil {
		t.Fatalf("expected hash to match default password: %v", err)
	}
}
