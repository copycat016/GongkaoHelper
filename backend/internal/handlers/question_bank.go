package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type QuestionBankHandler struct {
	service *services.QuestionBankService
}

func NewQuestionBankHandler(service *services.QuestionBankService) *QuestionBankHandler {
	return &QuestionBankHandler{service: service}
}

func (h *QuestionBankHandler) List(c *gin.Context) {
	items, err := h.service.List(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50091, "list question bank failed")
		return
	}
	response.Success(c, items)
}

func (h *QuestionBankHandler) Get(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40091, "invalid question id")
		return
	}

	item, err := h.service.Get(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "get question failed")
		return
	}
	response.Success(c, item)
}

func (h *QuestionBankHandler) Delete(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40091, "invalid question id")
		return
	}

	if err := h.service.Delete(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete question failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *QuestionBankHandler) Update(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40091, "invalid question id")
		return
	}

	var payload services.QuestionBankItem
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40092, "invalid question payload")
		return
	}

	item, err := h.service.Update(userIDFromRequest(c), id, payload)
	if err != nil {
		writeServiceError(c, err, "update question failed")
		return
	}
	response.Success(c, item)
}
