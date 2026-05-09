package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/database"
	"gkweb/backend/internal/response"
)

type DBHandler struct {
	db *gorm.DB
}

func NewDBHandler(db *gorm.DB) *DBHandler {
	return &DBHandler{db: db}
}

func (h *DBHandler) Ping(c *gin.Context) {
	if err := database.Ping(c.Request.Context(), h.db); err != nil {
		response.Error(c, http.StatusServiceUnavailable, 50301, "database connection failed")
		return
	}

	response.Success(c, gin.H{
		"status":  "ok",
		"message": "database connection successful",
	})
}
