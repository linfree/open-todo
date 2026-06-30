package api

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterBackupRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/backup", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "backup not implemented yet"})
	})
}
