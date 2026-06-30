package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
)

type CategoryHandler struct {
	DB database.Database
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	userID := c.GetString("user_id")
	cats, err := h.DB.GetCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *CategoryHandler) SaveCategory(c *gin.Context) {
	userID := c.GetString("user_id")
	var cat database.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat.UserID = userID
	if err := h.DB.SaveCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
