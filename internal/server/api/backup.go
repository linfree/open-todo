package api

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/app"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterBackupRoutes(r *gin.RouterGroup, st *store.Store) {
	r.POST("/backup", func(c *gin.Context) {
		var cfg app.WebDAVConfig
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		filename, err := app.BackupToWebDAV(st, cfg)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"filename": filename})
	})

	r.POST("/restore", func(c *gin.Context) {
		var req struct {
			Config   app.WebDAVConfig `json:"config"`
			Filename string           `json:"filename"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := app.RestoreFromWebDAV(st, req.Config, req.Filename); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})
}
