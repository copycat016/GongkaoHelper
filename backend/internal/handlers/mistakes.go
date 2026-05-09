package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type MistakeHandler struct {
	service *services.MistakeService
}

func NewMistakeHandler(service *services.MistakeService) *MistakeHandler {
	return &MistakeHandler{service: service}
}

func (h *MistakeHandler) List(c *gin.Context) {
	filters := services.MistakeFilters{
		Subject:      c.Query("subject"),
		QuestionType: c.Query("question_type"),
		ErrorReason:  c.Query("error_reason"),
		Mastery:      c.Query("mastery"),
		Tag:          c.Query("tag"),
	}

	mistakes, err := h.service.List(userIDFromRequest(c), filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50021, "list mistakes failed")
		return
	}

	response.Success(c, mistakes)
}

func (h *MistakeHandler) Create(c *gin.Context) {
	var mistake models.Mistake
	if err := c.ShouldBindJSON(&mistake); err != nil {
		response.Error(c, http.StatusBadRequest, 40021, "invalid mistake payload")
		return
	}

	mistake.UserID = userIDFromRequest(c)
	if err := h.service.Create(&mistake); err != nil {
		response.Error(c, http.StatusInternalServerError, 50022, "create mistake failed")
		return
	}

	response.Success(c, mistake)
}

func (h *MistakeHandler) Get(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40022, "invalid mistake id")
		return
	}

	mistake, err := h.service.Get(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "get mistake failed")
		return
	}

	response.Success(c, mistake)
}

func (h *MistakeHandler) Update(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40022, "invalid mistake id")
		return
	}

	var payload models.Mistake
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40021, "invalid mistake payload")
		return
	}

	mistake, err := h.service.Update(userIDFromRequest(c), id, &payload)
	if err != nil {
		writeServiceError(c, err, "update mistake failed")
		return
	}

	response.Success(c, mistake)
}

func (h *MistakeHandler) Delete(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40022, "invalid mistake id")
		return
	}

	if err := h.service.Delete(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete mistake failed")
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

func (h *MistakeHandler) Review(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40022, "invalid mistake id")
		return
	}

	var payload services.MistakeReviewInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40023, "invalid review payload")
		return
	}

	mistake, err := h.service.Review(userIDFromRequest(c), id, payload)
	if err != nil {
		writeServiceError(c, err, "review mistake failed")
		return
	}

	response.Success(c, mistake)
}
