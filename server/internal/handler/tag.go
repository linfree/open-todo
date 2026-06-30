package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
)

type TagHandler struct {
	DB database.Database
}

func (h *TagHandler) GetTags(c *gin.Context) {
	userID := c.GetString("user_id")
	tags, err := h.DB.GetTags(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) SaveTag(c *gin.Context) {
	userID := c.GetString("user_id")
	var tag database.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag.UserID = userID
	if err := h.DB.SaveTag(&tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.DeleteTag(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
