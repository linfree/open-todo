package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port         int    `json:"port"`
	ServerURL    string `json:"server_url"`
	AuthToken    string `json:"auth_token"`
	AutoSync     bool   `json:"auto_sync"`
	DatabasePath string `json:"database_path"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Port:         18080,
		AutoSync:     false,
		DatabasePath: filepath.Join(home, ".open-todo", "open-todo.db"),
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".open-todo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	cfg := Default()
	path, err := configPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
