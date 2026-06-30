package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
)

type SyncHandler struct {
	DB database.Database
}

func (h *SyncHandler) Push(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Changes []database.SyncRecord `json:"changes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.StoreChanges(userID, req.Changes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"server_time": time.Now().UTC().Format(time.RFC3339)})
}

func (h *SyncHandler) Pull(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Since string `json:"since"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	since, _ := time.Parse(time.RFC3339, req.Since)
	if req.Since == "" {
		since = time.Unix(0, 0)
	}

	changes, err := h.DB.GetChangesSince(userID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changes": changes, "server_time": time.Now().UTC().Format(time.RFC3339)})
}
