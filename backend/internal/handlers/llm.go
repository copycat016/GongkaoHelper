package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type LLMHandler struct {
	service *services.LLMService
}

func NewLLMHandler(service *services.LLMService) *LLMHandler {
	return &LLMHandler{service: service}
}

func (h *LLMHandler) ListProviders(c *gin.Context) {
	providers, err := h.service.ListProviders(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "list providers failed")
		return
	}
	response.Success(c, providers)
}

func (h *LLMHandler) CreateProvider(c *gin.Context) {
	var provider models.LLMProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid provider payload")
		return
	}

	provider.UserID = userIDFromRequest(c)
	if err := h.service.CreateProvider(&provider); err != nil {
		response.Error(c, http.StatusInternalServerError, 50002, "create provider failed")
		return
	}
	response.Success(c, provider)
}

func (h *LLMHandler) UpdateProvider(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40002, "invalid provider id")
		return
	}

	var payload models.LLMProvider
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid provider payload")
		return
	}

	provider, err := h.service.UpdateProvider(userIDFromRequest(c), id, &payload)
	if err != nil {
		writeServiceError(c, err, "update provider failed")
		return
	}
	response.Success(c, provider)
}

func (h *LLMHandler) DeleteProvider(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40002, "invalid provider id")
		return
	}

	if err := h.service.DeleteProvider(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete provider failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *LLMHandler) FetchProviderModels(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40002, "invalid provider id")
		return
	}

	modelsList, err := h.service.FetchProviderModels(userIDFromRequest(c), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeServiceError(c, err, "provider not found")
			return
		}
		response.Error(c, http.StatusBadGateway, 50201, err.Error())
		return
	}
	response.Success(c, modelsList)
}

func (h *LLMHandler) ListModels(c *gin.Context) {
	modelsList, err := h.service.ListModels(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50003, "list models failed")
		return
	}
	response.Success(c, modelsList)
}

func (h *LLMHandler) CreateModel(c *gin.Context) {
	var model models.LLMModel
	if err := c.ShouldBindJSON(&model); err != nil {
		response.Error(c, http.StatusBadRequest, 40003, "invalid model payload")
		return
	}

	model.UserID = userIDFromRequest(c)
	if err := h.service.CreateModel(&model); err != nil {
		response.Error(c, http.StatusBadRequest, 40004, err.Error())
		return
	}
	response.Success(c, model)
}

func (h *LLMHandler) UpdateModel(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40005, "invalid model id")
		return
	}

	var payload models.LLMModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40003, "invalid model payload")
		return
	}

	model, err := h.service.UpdateModel(userIDFromRequest(c), id, &payload)
	if err != nil {
		writeServiceError(c, err, "update model failed")
		return
	}
	response.Success(c, model)
}

func (h *LLMHandler) DeleteModel(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40005, "invalid model id")
		return
	}

	if err := h.service.DeleteModel(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete model failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func writeServiceError(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, 40401, "resource not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, 50004, message)
}
