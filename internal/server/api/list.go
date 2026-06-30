package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterListRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/lists", func(c *gin.Context) {
		lists, err := st.GetLists()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, lists)
	})

	r.POST("/lists", func(c *gin.Context) {
		var l store.TaskList
		if err := c.ShouldBindJSON(&l); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveList(&l); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, l)
	})

	r.DELETE("/lists/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := st.DeleteList(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
