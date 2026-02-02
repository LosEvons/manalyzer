package manalyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents persistent application configuration.
type Config struct {
	Version      int            `json:"version"`
	Players      []PlayerConfig `json:"players"`
	LastDemoPath string         `json:"lastDemoPath,omitempty"`
	Preferences  Preferences    `json:"preferences"`
}

// PlayerConfig stores a player's name and SteamID64.
type PlayerConfig struct {
	Name      string `json:"name"`
	SteamID64 string `json:"steamID64"`
}

// Preferences stores UI preferences.
type Preferences struct {
	AutoSave bool `json:"autoSave"`
}

// DefaultConfig returns a config with defaults.
func DefaultConfig() *Config {
	return &Config{
		Version:     1,
		Players:     make([]PlayerConfig, 0),
		Preferences: Preferences{AutoSave: true},
	}
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".manalyzer", "config.json"), nil
	}
	return filepath.Join(configDir, "manalyzer", "config.json"), nil
}

// LoadConfig loads configuration from disk.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), nil
	}

	return &config, nil
}

// SaveConfig saves configuration to disk.
func SaveConfig(config *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
