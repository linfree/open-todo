package api

import (
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/app"
	"github.com/linfree/open-todo/internal/config"
)

// RegisterAIRoutes registers AI configuration and service endpoints.
func RegisterAIRoutes(r *gin.RouterGroup) {
	// AI config endpoints (existing)
	r.GET("/ai-config", handleGetAIConfig)
	r.POST("/ai-config", handleSaveAIConfig)

	// AI service endpoints (new)
	aiGroup := r.Group("/ai")
	aiGroup.GET("/status", handleAIStatus)
	aiGroup.POST("/test", handleAITest)
	aiGroup.POST("/parse-task", handleParseTask)
	aiGroup.POST("/breakdown-task", handleBreakdownTask)
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
	if len(apiKey) > 0 && strings.Contains(apiKey, "*") {
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

// handleAIStatus returns whether AI is enabled and configured.
func handleAIStatus(c *gin.Context) {
	svc := app.NewAIService()
	enabled, configured, model := svc.Status()
	c.JSON(200, gin.H{
		"enabled":    enabled,
		"configured": configured,
		"model":      model,
	})
}

// handleAITest tests the AI API connection.
func handleAITest(c *gin.Context) {
	svc := app.NewAIService()
	if err := svc.TestConnection(); err != nil {
		c.JSON(200, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// handleParseTask parses natural language input into a structured task.
func handleParseTask(c *gin.Context) {
	var req struct {
		Input string `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Input == "" {
		c.JSON(400, gin.H{"error": "input不能为空"})
		return
	}

	svc := app.NewAIService()
	result, err := svc.ParseTask(req.Input)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// handleBreakdownTask breaks down a task into subtasks using AI.
func handleBreakdownTask(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(400, gin.H{"error": "title不能为空"})
		return
	}

	svc := app.NewAIService()
	result, err := svc.BreakdownTask(req.Title, req.Description)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}
