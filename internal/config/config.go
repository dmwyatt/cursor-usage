package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// Config holds the CLI configuration persisted to disk.
type Config struct {
	SessionToken string `json:"session_token"`
	BaseURL      string `json:"base_url,omitempty"`
}

// DefaultPath returns the platform-appropriate config file path.
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, "cursor-usage", "config.json")
}

// Load reads the config from the default path.
// Returns a zero-value Config and nil error if the file does not exist.
func Load() (Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom reads the config from a specific path.
// Returns a zero-value Config and nil error if the file does not exist.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the config to the default path.
func Save(cfg Config) error {
	return SaveTo(DefaultPath(), cfg)
}

// SaveTo writes the config to a specific path, creating parent directories as needed.
func SaveTo(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
