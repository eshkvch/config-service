package config

import (
	"strings"
	"testing"
)

func TestLoadUsesDatabaseURLAndPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/configs?sslmode=disable")
	t.Setenv("PORT", "9090")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.DSN != "postgres://user:pass@db:5432/configs?sslmode=disable" {
		t.Fatalf("unexpected DSN: %s", cfg.Database.DSN)
	}
	if cfg.HTTP.Port != "9090" {
		t.Fatalf("unexpected port: %s", cfg.HTTP.Port)
	}
}

func TestLoadBuildsDatabaseDSNFromParts(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "configdb")
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "15432")
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "postgres://app:secret@postgres:15432/configdb?sslmode=disable"
	if cfg.Database.DSN != want {
		t.Fatalf("DSN = %q, want %q", cfg.Database.DSN, want)
	}
	if cfg.HTTP.Port != "8080" {
		t.Fatalf("port = %q, want default 8080", cfg.HTTP.Port)
	}
}

func TestLoadReturnsErrorWhenDatabaseConfigMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "set DATABASE_URL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatabaseDSNUsesDefaultHostAndPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "configdb")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")

	dsn, err := databaseDSN()
	if err != nil {
		t.Fatalf("databaseDSN() error = %v", err)
	}

	want := "postgres://app:secret@localhost:5432/configdb?sslmode=disable"
	if dsn != want {
		t.Fatalf("DSN = %q, want %q", dsn, want)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Setenv("CONFIG_TEST_VALUE", "configured")

	if got := getEnvOrDefault("CONFIG_TEST_VALUE", "default"); got != "configured" {
		t.Fatalf("getEnvOrDefault() = %q", got)
	}
	if got := getEnvOrDefault("CONFIG_TEST_MISSING", "default"); got != "default" {
		t.Fatalf("getEnvOrDefault() default = %q", got)
	}
}

func TestFirstEnv(t *testing.T) {
	t.Setenv("FIRST_ENV_A", "")
	t.Setenv("FIRST_ENV_B", "second")
	t.Setenv("FIRST_ENV_C", "third")

	if got := firstEnv("FIRST_ENV_A", "FIRST_ENV_B", "FIRST_ENV_C"); got != "second" {
		t.Fatalf("firstEnv() = %q", got)
	}
	if got := firstEnv("FIRST_ENV_UNKNOWN"); got != "" {
		t.Fatalf("firstEnv() for missing key = %q", got)
	}
}
