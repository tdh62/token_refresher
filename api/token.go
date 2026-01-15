package api

import (
	"jwt_refresher/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	db *database.DB
}

func NewTokenHandler(db *database.DB) *TokenHandler {
	return &TokenHandler{db: db}
}

// GetToken 获取当前有效token
func (h *TokenHandler) GetToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.db.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  project.CurrentAccessToken,
		"refresh_token": project.CurrentRefreshToken,
		"expires_at":    project.TokenExpiresAt,
	})
}

// GetLogs 获取刷新日志
func (h *TokenHandler) GetLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.db.GetProjectLogs(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
