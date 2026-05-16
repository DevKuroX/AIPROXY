package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save and clear env to test defaults
	oldPort := os.Getenv("PORT")
	os.Unsetenv("PORT")
	defer os.Setenv("PORT", oldPort)

	cfg := Load()
	if cfg.Port == 0 {
		t.Fatal("port should not be 0")
	}
	t.Logf("Default port: %d", cfg.Port)
}

func TestLoadCustomPort(t *testing.T) {
	os.Setenv("PORT", "5555")
	defer os.Unsetenv("PORT")

	cfg := Load()
	if cfg.Port != 5555 {
		t.Fatalf("expected 5555, got %d", cfg.Port)
	}
}
