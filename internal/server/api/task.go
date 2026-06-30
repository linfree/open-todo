package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterTaskRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/tasks", func(c *gin.Context) {
		tasks, err := st.GetTasks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	r.POST("/tasks", func(c *gin.Context) {
		var t store.Task
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveTask(&t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, t)
	})

	r.DELETE("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := st.SoftDeleteTask(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
