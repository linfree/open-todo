package api

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/config"
)

// RegisterAIRoutes registers AI configuration endpoints.
func RegisterAIRoutes(r *gin.RouterGroup) {
	r.GET("/ai-config", handleGetAIConfig)
	r.POST("/ai-config", handleSaveAIConfig)
}

// maskedKey returns the API key with only the last 4 characters visible.
func maskedKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	masked := ""
	for range len(key) - 4 {
		masked += "*"
	}
	return masked + key[len(key)-4:]
}

func handleGetAIConfig(c *gin.Context) {
	cfg, err := config.LoadAIConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Return config with masked API key
	c.JSON(200, gin.H{
		"enabled":       cfg.Enabled,
		"base_url":      cfg.BaseURL,
		"api_key":       maskedKey(cfg.APIKey),
		"model":         cfg.Model,
		"system_prompt": cfg.SystemPrompt,
	})
}

func handleSaveAIConfig(c *gin.Context) {
	var req struct {
		Enabled      bool   `json:"enabled"`
		BaseURL      string `json:"base_url"`
		APIKey       string `json:"api_key"`
		Model        string `json:"model"`
		SystemPrompt string `json:"system_prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Load existing config to preserve the API key if the incoming one is masked
	existing, err := config.LoadAIConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	apiKey := req.APIKey
	// If the incoming key is masked (contains ****), keep the existing key
	if len(apiKey) > 0 && apiKey[:4] == "****" {
		apiKey = existing.APIKey
	}

	cfg := &config.AIConfig{
		Enabled:      req.Enabled,
		BaseURL:      req.BaseURL,
		APIKey:       apiKey,
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
	}

	if err := config.SaveAIConfig(cfg); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"ok":      true,
		"api_key": maskedKey(apiKey),
	})
}
