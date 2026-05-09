package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type OCRHandler struct {
	service *services.BaiduOCRService
}

func NewOCRHandler(service *services.BaiduOCRService) *OCRHandler {
	return &OCRHandler{service: service}
}

func (h *OCRHandler) Engines(c *gin.Context) {
	response.Success(c, h.service.Engines())
}

func (h *OCRHandler) Scenes(c *gin.Context) {
	response.Success(c, h.service.Scenes())
}

func (h *OCRHandler) Config(c *gin.Context) {
	response.Success(c, h.service.PublicConfig())
}

func (h *OCRHandler) UpdateConfig(c *gin.Context) {
	var payload services.OCRServerConfig
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40063, "invalid ocr config payload")
		return
	}
	cfg, err := h.service.UpdateConfig(payload)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50062, "save ocr config failed")
		return
	}
	response.Success(c, cfg)
}

func (h *OCRHandler) Recognize(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40061, "ocr file is required")
		return
	}

	result, err := h.service.Recognize(userIDFromRequest(c), c.PostForm("scene"), c.PostForm("engine"), file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40062, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *OCRHandler) MonthUsage(c *gin.Context) {
	usage, err := h.service.MonthUsage(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50061, "get ocr usage failed")
		return
	}
	response.Success(c, usage)
}
