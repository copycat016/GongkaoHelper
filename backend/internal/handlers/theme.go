package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type ThemeHandler struct {
	service *services.ThemeService
}

func NewThemeHandler(service *services.ThemeService) *ThemeHandler {
	return &ThemeHandler{service: service}
}

func (h *ThemeHandler) Get(c *gin.Context) {
	config, err := h.service.Get(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50031, "get theme config failed")
		return
	}
	response.Success(c, config)
}

func (h *ThemeHandler) Save(c *gin.Context) {
	var config models.ThemeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, http.StatusBadRequest, 40031, "invalid theme config payload")
		return
	}

	result, err := h.service.Save(userIDFromRequest(c), &config)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50032, "save theme config failed")
		return
	}
	response.Success(c, result)
}
