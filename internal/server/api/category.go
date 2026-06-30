package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterCategoryRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/categories", func(c *gin.Context) {
		cats, err := st.GetCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cats)
	})
	r.POST("/categories", func(c *gin.Context) {
		var cat store.Category
		if err := c.ShouldBindJSON(&cat); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveCategory(&cat); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cat)
	})
	r.DELETE("/categories/:id", func(c *gin.Context) {
		if err := st.DeleteCategory(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
