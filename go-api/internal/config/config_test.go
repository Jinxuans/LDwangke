package config

import "testing"

func TestConfigSecurityDefaults(t *testing.T) {
	cfg := &Config{}

	if !cfg.AllowLegacyPlaintextPasswords() {
		t.Fatalf("expected legacy plaintext passwords to be allowed by default")
	}
	if cfg.CreateDefaultAdminEnabled() {
		t.Fatalf("expected default admin bootstrap to be disabled by default")
	}
	if !cfg.AutoMigrateEnabled() {
		t.Fatalf("expected auto migration to be enabled by default")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("GO_API_SERVER_PORT", "9090")
	t.Setenv("GO_API_DATABASE_PORT", "3307")
	t.Setenv("GO_API_SECURITY_ALLOW_LEGACY_PLAINTEXT_PASSWORDS", "false")
	t.Setenv("GO_API_BOOTSTRAP_CREATE_DEFAULT_ADMIN", "true")
	t.Setenv("GO_API_BOOTSTRAP_AUTO_MIGRATE", "false")
	t.Setenv("GO_API_BOOTSTRAP_MIGRATIONS_DIR", "custom-migrations")

	cfg := &Config{}
	applyEnvOverrides(cfg)

	if cfg.Server.Port != "9090" {
		t.Fatalf("expected server port override, got %q", cfg.Server.Port)
	}
	if cfg.Database.Port != 3307 {
		t.Fatalf("expected database port override, got %d", cfg.Database.Port)
	}
	if cfg.AllowLegacyPlaintextPasswords() {
		t.Fatalf("expected legacy plaintext passwords to be disabled by env override")
	}
	if !cfg.CreateDefaultAdminEnabled() {
		t.Fatalf("expected default admin bootstrap to be enabled by env override")
	}
	if cfg.AutoMigrateEnabled() {
		t.Fatalf("expected auto migration to be disabled by env override")
	}
	if cfg.MigrationsDirValue() != "custom-migrations" {
		t.Fatalf("expected migrations dir override, got %q", cfg.MigrationsDirValue())
	}
}
