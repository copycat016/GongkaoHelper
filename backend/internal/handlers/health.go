package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/version"
)

type HealthHandler struct {
	startedAt time.Time
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{startedAt: time.Now()}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status":     "ok",
		"service":    "gkweb-backend",
		"version":    version.Version,
		"channel":    version.Channel,
		"commit":     version.Commit,
		"started_at": h.startedAt.Format(time.RFC3339),
	})
}
