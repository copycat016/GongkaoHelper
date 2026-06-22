package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"gkweb/backend/internal/response"
	"gkweb/backend/internal/services"
)

type DailyTaskHandler struct {
	service *services.DailyTaskService
}

func NewDailyTaskHandler(service *services.DailyTaskService) *DailyTaskHandler {
	return &DailyTaskHandler{service: service}
}

// List GET /api/tasks?date=YYYY-MM-DD
func (h *DailyTaskHandler) List(c *gin.Context) {
	tasks, err := h.service.List(userIDFromRequest(c), c.Query("date"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40051, err.Error())
		return
	}
	response.Success(c, tasks)
}

// Summary GET /api/tasks/summary?date=YYYY-MM-DD
func (h *DailyTaskHandler) Summary(c *gin.Context) {
	summary, err := h.service.Summary(userIDFromRequest(c), c.Query("date"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40052, err.Error())
		return
	}
	response.Success(c, summary)
}

// Create POST /api/tasks
func (h *DailyTaskHandler) Create(c *gin.Context) {
	var input services.DailyTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40053, "invalid task payload")
		return
	}
	task, err := h.service.Create(userIDFromRequest(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, err.Error())
		return
	}
	response.Success(c, task)
}

// Update PUT /api/tasks/:id
func (h *DailyTaskHandler) Update(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid task id")
		return
	}
	var input services.DailyTaskInput
	touched, err := bindJSONWithTouchedFields(c, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40053, "invalid task payload")
		return
	}
	input.Touched = touched
	task, err := h.service.Update(userIDFromRequest(c), id, input)
	if err != nil {
		writeServiceError(c, err, "update task failed")
		return
	}
	response.Success(c, task)
}

// Toggle POST /api/tasks/:id/toggle
func (h *DailyTaskHandler) Toggle(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid task id")
		return
	}
	task, err := h.service.Toggle(userIDFromRequest(c), id)
	if err != nil {
		writeServiceError(c, err, "toggle task failed")
		return
	}
	response.Success(c, task)
}

// Delete DELETE /api/tasks/:id
func (h *DailyTaskHandler) Delete(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, "invalid task id")
		return
	}
	if err := h.service.Delete(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete task failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *DailyTaskHandler) ListDailyTasks(c *gin.Context) {
	tasks, err := h.service.ListDailyTasks(userIDFromRequest(c), c.Query("date"), c.Query("unscheduled") == "true", c.Query("weekly_task_id"), c.Query("deadline"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40056, err.Error())
		return
	}
	response.Success(c, tasks)
}

func (h *DailyTaskHandler) CreateDailyTask(c *gin.Context) {
	var input services.DailyTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40053, "invalid task payload")
		return
	}
	task, err := h.service.CreateDailyTask(userIDFromRequest(c), input, false)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, err.Error())
		return
	}
	response.Success(c, task)
}

func (h *DailyTaskHandler) ListStageGoals(c *gin.Context) {
	goals, err := h.service.ListStageGoals(userIDFromRequest(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50051, "list stage goals failed")
		return
	}
	response.Success(c, goals)
}

func (h *DailyTaskHandler) CreateStageGoal(c *gin.Context) {
	var input services.StageGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40057, "invalid stage goal payload")
		return
	}
	goal, err := h.service.CreateStageGoal(userIDFromRequest(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40058, err.Error())
		return
	}
	response.Success(c, goal)
}

func (h *DailyTaskHandler) UpdateStageGoal(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40059, "invalid stage goal id")
		return
	}
	var input services.StageGoalInput
	touched, err := bindJSONWithTouchedFields(c, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40057, "invalid stage goal payload")
		return
	}
	input.Touched = touched
	goal, err := h.service.UpdateStageGoal(userIDFromRequest(c), id, input)
	if err != nil {
		writeServiceError(c, err, "update stage goal failed")
		return
	}
	response.Success(c, goal)
}

func (h *DailyTaskHandler) DeleteStageGoal(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40059, "invalid stage goal id")
		return
	}
	if err := h.service.DeleteStageGoal(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete stage goal failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *DailyTaskHandler) ListStageItems(c *gin.Context) {
	items, err := h.service.ListStageItems(userIDFromRequest(c), c.Query("stage_goal_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40065, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *DailyTaskHandler) CreateStageItem(c *gin.Context) {
	var input services.StageItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40066, "invalid stage item payload")
		return
	}
	item, err := h.service.CreateStageItem(userIDFromRequest(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40067, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *DailyTaskHandler) UpdateStageItem(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40068, "invalid stage item id")
		return
	}
	var input services.StageItemInput
	touched, err := bindJSONWithTouchedFields(c, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40066, "invalid stage item payload")
		return
	}
	input.Touched = touched
	item, err := h.service.UpdateStageItem(userIDFromRequest(c), id, input)
	if err != nil {
		writeServiceError(c, err, "update stage item failed")
		return
	}
	response.Success(c, item)
}

func (h *DailyTaskHandler) DeleteStageItem(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40068, "invalid stage item id")
		return
	}
	if err := h.service.DeleteStageItem(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete stage item failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *DailyTaskHandler) ListWeeklyTasks(c *gin.Context) {
	tasks, err := h.service.ListWeeklyTasks(userIDFromRequest(c), c.Query("week_start"), c.Query("stage_goal_id"), c.Query("stage_item_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40060, err.Error())
		return
	}
	response.Success(c, tasks)
}

func (h *DailyTaskHandler) CreateWeeklyTask(c *gin.Context) {
	var input services.WeeklyTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40061, "invalid weekly task payload")
		return
	}
	task, err := h.service.CreateWeeklyTask(userIDFromRequest(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40062, err.Error())
		return
	}
	response.Success(c, task)
}

func (h *DailyTaskHandler) UpdateWeeklyTask(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40063, "invalid weekly task id")
		return
	}
	var input services.WeeklyTaskInput
	touched, err := bindJSONWithTouchedFields(c, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40061, "invalid weekly task payload")
		return
	}
	input.Touched = touched
	task, err := h.service.UpdateWeeklyTask(userIDFromRequest(c), id, input)
	if err != nil {
		writeServiceError(c, err, "update weekly task failed")
		return
	}
	response.Success(c, task)
}

func (h *DailyTaskHandler) MaterializeWeeklyTask(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40063, "invalid weekly task id")
		return
	}
	var input services.WeeklyTaskMaterializeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40064, "invalid weekly task materialize payload")
		return
	}
	result, err := h.service.MaterializeWeeklyTask(userIDFromRequest(c), id, input)
	if err != nil {
		writeServiceError(c, err, "materialize weekly task failed")
		return
	}
	response.Success(c, result)
}

func bindJSONWithTouchedFields(c *gin.Context, dest any) (map[string]bool, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	touched := make(map[string]bool, len(raw))
	for key := range raw {
		touched[key] = true
	}
	return touched, nil
}

func (h *DailyTaskHandler) DeleteWeeklyTask(c *gin.Context) {
	id, err := uintParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40063, "invalid weekly task id")
		return
	}
	if err := h.service.DeleteWeeklyTask(userIDFromRequest(c), id); err != nil {
		writeServiceError(c, err, "delete weekly task failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
