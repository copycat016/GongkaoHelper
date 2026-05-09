package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type PromptHandler struct {
	service *services.PromptService
}

func NewPromptHandler(service *services.PromptService) *PromptHandler {
	return &PromptHandler{service: service}
}

func (h *PromptHandler) List(c *gin.Context) {
	prompts, err := h.service.List(userIDFromRequest(c), c.Query("task_type"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50011, "list prompts failed")
		return
	}
	response.Success(c, prompts)
}

func (h *PromptHandler) Create(c *gin.Context) {
	var prompt models.PromptTemplate
	if err := c.ShouldBindJSON(&prompt); err != nil {
		response.Error(c, http.StatusBadRequest, 40011, "invalid prompt payload")
		return
	}

	prompt.UserID = userIDFromRequest(c)
	if err := h.service.Create(&prompt); err != nil {
		response.Error(c, http.StatusInternalServerError, 50012, "create prompt failed")
		return
	}
	response.Success(c, prompt)
}

func (h *PromptHandler) Update(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40012, "invalid prompt id")
		return
	}

	var payload models.PromptTemplate
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40011, "invalid prompt payload")
		return
	}

	prompt, err := h.service.Update(userIDFromRequest(c), id, &payload)
	if err != nil {
		writeServiceError(c, err, "update prompt failed")
		return
	}
	response.Success(c, prompt)
}

func (h *PromptHandler) Delete(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40012, "invalid prompt id")
		return
	}

	if err := h.service.Delete(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete prompt failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
