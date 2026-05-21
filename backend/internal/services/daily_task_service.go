package services

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type DailyTaskService struct {
	db *gorm.DB
}

func NewDailyTaskService(db *gorm.DB) *DailyTaskService {
	return &DailyTaskService{db: db}
}

type StageGoalInput struct {
	Title        string `json:"title"`
	Note         string `json:"note"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	TargetMetric string `json:"target_metric"`
	Status       string `json:"status"`
	SortOrder    *int   `json:"sort_order"`
}

type WeeklyTaskInput struct {
	StageGoalID   *uint  `json:"stage_goal_id"`
	Title         string `json:"title"`
	Note          string `json:"note"`
	WeekStart     string `json:"week_start"`
	DueDate       string `json:"due_date"`
	Priority      string `json:"priority"`
	Status        string `json:"status"`
	TargetMinutes int    `json:"target_minutes"`
	SortOrder     *int   `json:"sort_order"`
}

type DailyTaskInput struct {
	StageGoalID  *uint  `json:"stage_goal_id"`
	WeeklyTaskID *uint  `json:"weekly_task_id"`
	Title        string `json:"title"`
	Note         string `json:"note"`
	PlanDate     string `json:"plan_date"`
	DueDate      string `json:"due_date"`
	Priority     string `json:"priority"`
	Done         *bool  `json:"done"`
	SortOrder    *int   `json:"sort_order"`
}

type DailySummary struct {
	Date         string `json:"date"`
	TotalCount   int    `json:"total_count"`
	DoneCount    int    `json:"done_count"`
	PendingCount int    `json:"pending_count"`
	OverdueCount int    `json:"overdue_count"`
}

type WeeklyTaskView struct {
	models.WeeklyTask
	DailyTotal      int `json:"daily_total"`
	DailyDone       int `json:"daily_done"`
	ProgressPercent int `json:"progress_percent"`
}

type StageGoalView struct {
	models.StageGoal
	WeeklyTotal     int `json:"weekly_total"`
	WeeklyDone      int `json:"weekly_done"`
	DailyTotal      int `json:"daily_total"`
	DailyDone       int `json:"daily_done"`
	ProgressPercent int `json:"progress_percent"`
}

func (s *DailyTaskService) List(userID uint, date string) ([]models.DailyTask, error) {
	day, err := parseLocalDay(date)
	if err != nil {
		return nil, err
	}
	end := day.AddDate(0, 0, 1)

	var tasks []models.DailyTask
	err = s.db.
		Where("user_id = ? AND plan_date >= ? AND plan_date < ?", userID, day, end).
		Order("done asc, sort_order asc, id asc").
		Find(&tasks).Error
	return tasks, err
}

func (s *DailyTaskService) Summary(userID uint, date string) (*DailySummary, error) {
	tasks, err := s.List(userID, date)
	if err != nil {
		return nil, err
	}
	day, _ := parseLocalDay(date)

	summary := &DailySummary{Date: day.Format("2006-01-02"), TotalCount: len(tasks)}
	for _, task := range tasks {
		if task.Done {
			summary.DoneCount++
		} else {
			summary.PendingCount++
		}
		if !task.Done && task.DueDate != nil && task.DueDate.Before(day) {
			summary.OverdueCount++
		}
	}
	return summary, nil
}

func (s *DailyTaskService) Create(userID uint, input DailyTaskInput) (*models.DailyTask, error) {
	return s.CreateDailyTask(userID, input, true)
}

func (s *DailyTaskService) CreateDailyTask(userID uint, input DailyTaskInput, defaultToday bool) (*models.DailyTask, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	planDate, err := parseOptionalLocalDay(input.PlanDate)
	if err != nil {
		return nil, err
	}
	if planDate == nil && defaultToday {
		day, _ := parseLocalDay("")
		planDate = &day
	}
	dueDate, err := parseOptionalLocalDay(input.DueDate)
	if err != nil {
		return nil, err
	}

	task := models.DailyTask{
		BaseModel:    models.BaseModel{UserID: userID},
		StageGoalID:  input.StageGoalID,
		WeeklyTaskID: input.WeeklyTaskID,
		Title:        title,
		Note:         strings.TrimSpace(input.Note),
		PlanDate:     planDate,
		DueDate:      dueDate,
		Priority:     normalizePriority(input.Priority),
	}
	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}
	if input.Done != nil {
		s.applyDoneState(&task, *input.Done)
	}

	return &task, s.db.Create(&task).Error
}

func (s *DailyTaskService) Update(userID, id uint, input DailyTaskInput) (*models.DailyTask, error) {
	task, err := s.getDaily(userID, id)
	if err != nil {
		return nil, err
	}

	if title := strings.TrimSpace(input.Title); title != "" {
		task.Title = title
	}
	task.Note = strings.TrimSpace(input.Note)
	if input.StageGoalID != nil {
		task.StageGoalID = input.StageGoalID
	}
	if input.WeeklyTaskID != nil {
		task.WeeklyTaskID = input.WeeklyTaskID
	}
	if input.PlanDate != "" {
		planDate, err := parseOptionalLocalDay(input.PlanDate)
		if err != nil {
			return nil, err
		}
		task.PlanDate = planDate
	}
	if input.DueDate != "" {
		dueDate, err := parseOptionalLocalDay(input.DueDate)
		if err != nil {
			return nil, err
		}
		task.DueDate = dueDate
	}
	if input.Priority != "" {
		task.Priority = normalizePriority(input.Priority)
	}
	if input.Done != nil {
		s.applyDoneState(task, *input.Done)
	}
	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}

	return task, s.db.Save(task).Error
}

func (s *DailyTaskService) Toggle(userID, id uint) (*models.DailyTask, error) {
	task, err := s.getDaily(userID, id)
	if err != nil {
		return nil, err
	}
	s.applyDoneState(task, !task.Done)
	return task, s.db.Save(task).Error
}

func (s *DailyTaskService) Delete(userID, id uint) error {
	return s.DeleteDailyTask(userID, id)
}

func (s *DailyTaskService) ListDailyTasks(userID uint, date string, unscheduled bool, weeklyTaskID string) ([]models.DailyTask, error) {
	var tasks []models.DailyTask
	query := s.db.Where("user_id = ?", userID)
	if weeklyTaskID != "" {
		id, err := strconv.ParseUint(weeklyTaskID, 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("weekly_task_id = ?", uint(id))
	} else if unscheduled {
		query = query.Where("plan_date IS NULL")
	} else if date != "" {
		day, err := parseLocalDay(date)
		if err != nil {
			return nil, err
		}
		query = query.Where("plan_date >= ? AND plan_date < ?", day, day.AddDate(0, 0, 1))
	}
	err := query.Order("done asc, due_date asc, sort_order asc, id asc").Find(&tasks).Error
	return tasks, err
}

func (s *DailyTaskService) DeleteDailyTask(userID, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.DailyTask{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *DailyTaskService) ListStageGoals(userID uint) ([]StageGoalView, error) {
	var goals []models.StageGoal
	if err := s.db.Where("user_id = ?", userID).Order("status asc, sort_order asc, end_date asc, id asc").Find(&goals).Error; err != nil {
		return nil, err
	}

	views := make([]StageGoalView, 0, len(goals))
	for _, goal := range goals {
		view := StageGoalView{StageGoal: goal}
		var weekly []models.WeeklyTask
		if err := s.db.Where("user_id = ? AND stage_goal_id = ?", userID, goal.ID).Find(&weekly).Error; err != nil {
			return nil, err
		}
		view.WeeklyTotal = len(weekly)
		for _, task := range weekly {
			if task.Status == "done" {
				view.WeeklyDone++
			}
		}
		var daily []models.DailyTask
		if err := s.db.Where("user_id = ? AND stage_goal_id = ?", userID, goal.ID).Find(&daily).Error; err != nil {
			return nil, err
		}
		view.DailyTotal = len(daily)
		for _, task := range daily {
			if task.Done {
				view.DailyDone++
			}
		}
		view.ProgressPercent = percent(view.DailyDone+view.WeeklyDone, view.DailyTotal+view.WeeklyTotal)
		views = append(views, view)
	}
	return views, nil
}

func (s *DailyTaskService) CreateStageGoal(userID uint, input StageGoalInput) (*models.StageGoal, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	start, err := parseOptionalLocalDay(input.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := parseOptionalLocalDay(input.EndDate)
	if err != nil {
		return nil, err
	}
	goal := models.StageGoal{
		BaseModel:    models.BaseModel{UserID: userID},
		Title:        title,
		Note:         strings.TrimSpace(input.Note),
		StartDate:    start,
		EndDate:      end,
		TargetMetric: strings.TrimSpace(input.TargetMetric),
		Status:       normalizeStatus(input.Status, "active"),
	}
	if input.SortOrder != nil {
		goal.SortOrder = *input.SortOrder
	}
	return &goal, s.db.Create(&goal).Error
}

func (s *DailyTaskService) UpdateStageGoal(userID, id uint, input StageGoalInput) (*models.StageGoal, error) {
	var goal models.StageGoal
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&goal).Error; err != nil {
		return nil, err
	}
	if title := strings.TrimSpace(input.Title); title != "" {
		goal.Title = title
	}
	goal.Note = strings.TrimSpace(input.Note)
	goal.TargetMetric = strings.TrimSpace(input.TargetMetric)
	if input.StartDate != "" {
		start, err := parseOptionalLocalDay(input.StartDate)
		if err != nil {
			return nil, err
		}
		goal.StartDate = start
	}
	if input.EndDate != "" {
		end, err := parseOptionalLocalDay(input.EndDate)
		if err != nil {
			return nil, err
		}
		goal.EndDate = end
	}
	if input.Status != "" {
		goal.Status = normalizeStatus(input.Status, "active")
	}
	if input.SortOrder != nil {
		goal.SortOrder = *input.SortOrder
	}
	return &goal, s.db.Save(&goal).Error
}

func (s *DailyTaskService) DeleteStageGoal(userID, id uint) error {
	return deleteByUserID[models.StageGoal](s.db, userID, id)
}

func (s *DailyTaskService) ListWeeklyTasks(userID uint, weekStart string, stageGoalID string) ([]WeeklyTaskView, error) {
	query := s.db.Where("user_id = ?", userID)
	if weekStart != "" {
		start, err := weekStartDate(weekStart)
		if err != nil {
			return nil, err
		}
		query = query.Where("week_start = ?", start)
	}
	if stageGoalID != "" {
		id, err := strconv.ParseUint(stageGoalID, 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("stage_goal_id = ?", uint(id))
	}

	var weekly []models.WeeklyTask
	if err := query.Order("status asc, priority asc, due_date asc, sort_order asc, id asc").Find(&weekly).Error; err != nil {
		return nil, err
	}
	views := make([]WeeklyTaskView, 0, len(weekly))
	for _, task := range weekly {
		view := WeeklyTaskView{WeeklyTask: task}
		var daily []models.DailyTask
		if err := s.db.Where("user_id = ? AND weekly_task_id = ?", userID, task.ID).Find(&daily).Error; err != nil {
			return nil, err
		}
		view.DailyTotal = len(daily)
		for _, item := range daily {
			if item.Done {
				view.DailyDone++
			}
		}
		view.ProgressPercent = percent(view.DailyDone, view.DailyTotal)
		views = append(views, view)
	}
	return views, nil
}

func (s *DailyTaskService) CreateWeeklyTask(userID uint, input WeeklyTaskInput) (*models.WeeklyTask, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	weekStart, err := weekStartDate(input.WeekStart)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseOptionalLocalDay(input.DueDate)
	if err != nil {
		return nil, err
	}
	task := models.WeeklyTask{
		BaseModel:     models.BaseModel{UserID: userID},
		StageGoalID:   input.StageGoalID,
		Title:         title,
		Note:          strings.TrimSpace(input.Note),
		WeekStart:     weekStart,
		DueDate:       dueDate,
		Priority:      normalizePriority(input.Priority),
		Status:        normalizeStatus(input.Status, "todo"),
		TargetMinutes: input.TargetMinutes,
	}
	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}
	return &task, s.db.Create(&task).Error
}

func (s *DailyTaskService) UpdateWeeklyTask(userID, id uint, input WeeklyTaskInput) (*models.WeeklyTask, error) {
	var task models.WeeklyTask
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&task).Error; err != nil {
		return nil, err
	}
	if title := strings.TrimSpace(input.Title); title != "" {
		task.Title = title
	}
	task.Note = strings.TrimSpace(input.Note)
	if input.StageGoalID != nil {
		task.StageGoalID = input.StageGoalID
	}
	if input.WeekStart != "" {
		weekStart, err := weekStartDate(input.WeekStart)
		if err != nil {
			return nil, err
		}
		task.WeekStart = weekStart
	}
	if input.DueDate != "" {
		dueDate, err := parseOptionalLocalDay(input.DueDate)
		if err != nil {
			return nil, err
		}
		task.DueDate = dueDate
	}
	if input.Priority != "" {
		task.Priority = normalizePriority(input.Priority)
	}
	if input.Status != "" {
		task.Status = normalizeStatus(input.Status, "todo")
	}
	task.TargetMinutes = input.TargetMinutes
	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}
	return &task, s.db.Save(&task).Error
}

func (s *DailyTaskService) DeleteWeeklyTask(userID, id uint) error {
	return deleteByUserID[models.WeeklyTask](s.db, userID, id)
}

func (s *DailyTaskService) applyDoneState(task *models.DailyTask, done bool) {
	task.Done = done
	if done {
		now := time.Now()
		task.DoneAt = &now
	} else {
		task.DoneAt = nil
	}
}

func (s *DailyTaskService) getDaily(userID, id uint) (*models.DailyTask, error) {
	var task models.DailyTask
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func deleteByUserID[T any](db *gorm.DB, userID uint, id uint) error {
	result := db.Where("user_id = ? AND id = ?", userID, id).Delete(new(T))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func normalizePriority(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return "low"
	case "high":
		return "high"
	default:
		return "normal"
	}
}

func normalizeStatus(value string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "todo", "doing", "done", "cancelled", "active", "archived":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return fallback
	}
}

func percent(done int, total int) int {
	if total <= 0 {
		return 0
	}
	return int(float64(done) / float64(total) * 100)
}

func parseLocalDay(date string) (time.Time, error) {
	if strings.TrimSpace(date) == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func parseOptionalLocalDay(date string) (*time.Time, error) {
	if strings.TrimSpace(date) == "" {
		return nil, nil
	}
	parsed, err := parseLocalDay(date)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func weekStartDate(date string) (time.Time, error) {
	day, err := parseLocalDay(date)
	if err != nil {
		return time.Time{}, err
	}
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).AddDate(0, 0, -(weekday - 1)), nil
}
