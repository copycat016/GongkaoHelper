package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type PomodoroHandler struct {
	service *services.PomodoroService
}

func NewPomodoroHandler(service *services.PomodoroService) *PomodoroHandler {
	return &PomodoroHandler{service: service}
}

func (h *PomodoroHandler) CreateSession(c *gin.Context) {
	var session models.PomodoroSession
	if err := c.ShouldBindJSON(&session); err != nil {
		response.Error(c, http.StatusBadRequest, 40031, "invalid pomodoro session payload")
		return
	}

	session.UserID = userIDFromRequest(c)
	if err := h.service.CreateSession(&session); err != nil {
		response.Error(c, http.StatusInternalServerError, 50031, "create pomodoro session failed")
		return
	}

	response.Success(c, session)
}

func (h *PomodoroHandler) TodayStats(c *gin.Context) {
	stats, err := h.service.TodayStats(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50032, "get pomodoro stats failed")
		return
	}

	response.Success(c, stats)
}
