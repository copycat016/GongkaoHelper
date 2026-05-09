package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type StudyHandler struct {
	service *services.StudyService
}

func NewStudyHandler(service *services.StudyService) *StudyHandler {
	return &StudyHandler{service: service}
}

func (h *StudyHandler) ListLogs(c *gin.Context) {
	logs, err := h.service.ListLogs(userIDFromRequest(c), c.Query("date"), c.DefaultQuery("scope", "day"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50041, "list logs failed")
		return
	}
	response.Success(c, logs)
}

func (h *StudyHandler) LogStats(c *gin.Context) {
	stats, err := h.service.LogStats(userIDFromRequest(c), c.Query("date"), c.DefaultQuery("scope", "day"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50042, "get log stats failed")
		return
	}
	response.Success(c, stats)
}

func (h *StudyHandler) CreateLog(c *gin.Context) {
	var log models.StudyLog
	if err := c.ShouldBindJSON(&log); err != nil {
		response.Error(c, http.StatusBadRequest, 40041, "invalid log payload")
		return
	}
	log.UserID = userIDFromRequest(c)
	if err := h.service.CreateLog(&log); err != nil {
		response.Error(c, http.StatusInternalServerError, 50043, "create log failed")
		return
	}
	response.Success(c, log)
}

func (h *StudyHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(userIDFromRequest(c), c.Query("plan_type"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50044, "list plans failed")
		return
	}
	response.Success(c, plans)
}

func (h *StudyHandler) CreatePlan(c *gin.Context) {
	var plan models.StudyPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		response.Error(c, http.StatusBadRequest, 40042, "invalid plan payload")
		return
	}
	plan.UserID = userIDFromRequest(c)
	if err := h.service.CreatePlan(&plan); err != nil {
		response.Error(c, http.StatusInternalServerError, 50045, "create plan failed")
		return
	}
	response.Success(c, plan)
}

func (h *StudyHandler) UpdatePlan(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40043, "invalid plan id")
		return
	}

	var payload models.StudyPlan
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, 40042, "invalid plan payload")
		return
	}

	plan, err := h.service.UpdatePlan(userIDFromRequest(c), id, &payload)
	if err != nil {
		writeServiceError(c, err, "update plan failed")
		return
	}
	response.Success(c, plan)
}

func (h *StudyHandler) DeletePlan(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40043, "invalid plan id")
		return
	}
	if err := h.service.DeletePlan(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete plan failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *StudyHandler) CompletePlan(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40043, "invalid plan id")
		return
	}
	plan, err := h.service.CompletePlan(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "complete plan failed")
		return
	}
	response.Success(c, plan)
}

func (h *StudyHandler) CalendarEvents(c *gin.Context) {
	events, err := h.service.CalendarEvents(userIDFromRequest(c), c.Query("month"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50046, "list calendar events failed")
		return
	}
	response.Success(c, events)
}
