package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bytepharoh/rideflow/pkg/config"
)

func TestLoadEnvSetsValuesFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "TRIP_HTTP_PORT=8081\nTRIP_LOG_LEVEL=\"debug\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := config.LoadEnv(path); err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	if got := os.Getenv("TRIP_HTTP_PORT"); got != "8081" {
		t.Fatalf("TRIP_HTTP_PORT = %q, want %q", got, "8081")
	}

	if got := os.Getenv("TRIP_LOG_LEVEL"); got != "debug" {
		t.Fatalf("TRIP_LOG_LEVEL = %q, want %q", got, "debug")
	}
}

func TestLoadEnvDoesNotOverrideExistingValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte("TRIP_HTTP_PORT=8081\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	t.Setenv("TRIP_HTTP_PORT", "9090")

	if err := config.LoadEnv(path); err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	if got := os.Getenv("TRIP_HTTP_PORT"); got != "9090" {
		t.Fatalf("TRIP_HTTP_PORT = %q, want %q", got, "9090")
	}
}

func TestLoadEnvReturnsNilForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist-rideflow.env")

	if err := config.LoadEnv(path); err != nil {
		t.Fatalf("LoadEnv() error = %v, want nil for missing file", err)
	}
}

func TestLoadEnvRejectsInvalidLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte("TRIP_HTTP_PORT\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := config.LoadEnv(path); err == nil {
		t.Fatal("LoadEnv() error = nil, want non-nil")
	}
}

func TestLoadEnvRejectsEmptyKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte("=somevalue\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := config.LoadEnv(path); err == nil {
		t.Fatal("LoadEnv() error = nil, want non-nil for empty key")
	}
}
