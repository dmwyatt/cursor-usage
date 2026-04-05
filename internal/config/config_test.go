package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNonexistentReturnsZeroConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.SessionToken != "" {
		t.Errorf("expected empty SessionToken, got %q", cfg.SessionToken)
	}
	if cfg.BaseURL != "" {
		t.Errorf("expected empty BaseURL, got %q", cfg.BaseURL)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	original := Config{
		SessionToken: "test-token-abc123",
		BaseURL:      "https://custom.example.com",
	}

	if err := SaveTo(path, original); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.SessionToken != original.SessionToken {
		t.Errorf("SessionToken mismatch: got %q, want %q", loaded.SessionToken, original.SessionToken)
	}
	if loaded.BaseURL != original.BaseURL {
		t.Errorf("BaseURL mismatch: got %q, want %q", loaded.BaseURL, original.BaseURL)
	}
}

func TestSaveCreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "config.json")

	err := SaveTo(path, Config{SessionToken: "tok"})
	if err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("parent dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected parent to be a directory")
	}
}

func TestLoadMalformedJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte("{not json}"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestDefaultPathIsNonEmpty(t *testing.T) {
	p := DefaultPath()
	if p == "" {
		t.Error("DefaultPath returned empty string")
	}
}
