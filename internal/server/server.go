package server

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/server/api"
	"github.com/linfree/open-todo/internal/store"
)

func New(st *store.Store, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	apiGroup := r.Group("/api/v1")
	api.RegisterTaskRoutes(apiGroup, st)
	api.RegisterListRoutes(apiGroup, st)
	api.RegisterBackupRoutes(apiGroup, st)
	api.RegisterCategoryRoutes(apiGroup, st)
	api.RegisterTagRoutes(apiGroup, st)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
