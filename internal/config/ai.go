package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AIConfig holds OpenAI-compatible AI configuration.
type AIConfig struct {
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
}

// DefaultAIConfig returns sensible defaults for AI configuration.
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		Enabled:      false,
		BaseURL:      "https://api.openai.com/v1",
		Model:        "gpt-4o",
		SystemPrompt: "你是一个待办事项助手，帮助用户管理任务、设置提醒和规划日程。",
	}
}

func aiConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ai_config.json"), nil
}

// LoadAIConfig reads the AI configuration from ~/.open-todo/ai_config.json.
// If the file does not exist, defaults are returned.
func LoadAIConfig() (*AIConfig, error) {
	cfg := DefaultAIConfig()
	path, err := aiConfigPath()
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

// SaveAIConfig writes the AI configuration to ~/.open-todo/ai_config.json.
func SaveAIConfig(cfg *AIConfig) error {
	path, err := aiConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
