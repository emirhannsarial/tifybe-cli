package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey string `json:"api_key"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tifybeDir := filepath.Join(homeDir, ".tifybe")
	if err := os.MkdirAll(tifybeDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(tifybeDir, "credentials.json"), nil
}

func SaveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Only readable by the user
	return os.WriteFile(path, data, 0600)
}

func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // Return empty config if not exists
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse credentials.json: %w", err)
	}

	return &cfg, nil
}

func ClearConfig() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
