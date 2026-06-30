package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterTagRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/tags", func(c *gin.Context) {
		tags, err := st.GetTags()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tags)
	})
	r.POST("/tags", func(c *gin.Context) {
		var t store.Tag
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveTag(&t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, t)
	})
	r.DELETE("/tags/:id", func(c *gin.Context) {
		if err := st.DeleteTag(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
