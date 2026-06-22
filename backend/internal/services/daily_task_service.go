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
	Title        string          `json:"title"`
	Note         string          `json:"note"`
	StartDate    *string         `json:"start_date"`
	EndDate      *string         `json:"end_date"`
	TargetText   string          `json:"target_text"`
	TargetMetric string          `json:"target_metric"` // 兼容旧前端字段
	Status       string          `json:"status"`
	ProgressMode string          `json:"progress_mode"`
	SortOrder    *int            `json:"sort_order"`
	Touched      map[string]bool `json:"-"`
}

type StageItemInput struct {
	StageGoalID  uint            `json:"stage_goal_id"`
	Title        string          `json:"title"`
	Note         string          `json:"note"`
	StartDate    *string         `json:"start_date"`
	EndDate      *string         `json:"end_date"`
	TargetText   string          `json:"target_text"`
	TargetMetric string          `json:"target_metric"` // 兼容旧前端字段
	Status       string          `json:"status"`
	ProgressMode string          `json:"progress_mode"`
	SortOrder    *int            `json:"sort_order"`
	Touched      map[string]bool `json:"-"`
}

type WeeklyTaskInput struct {
	StageGoalID   *uint           `json:"stage_goal_id"`
	StageItemID   *uint           `json:"stage_item_id"`
	TaskKind      string          `json:"task_kind"`
	Title         string          `json:"title"`
	Note          string          `json:"note"`
	WeekStart     string          `json:"week_start"`
	WeekEnd       *string         `json:"week_end"`
	StartDate     *string         `json:"start_date"`
	EndDate       *string         `json:"end_date"`
	Deadline      *string         `json:"deadline"`
	DueDate       *string         `json:"due_date"` // 兼容旧前端字段
	Priority      string          `json:"priority"`
	Status        string          `json:"status"`
	ExecuteMode   string          `json:"execute_mode"`
	TargetMinutes int             `json:"target_minutes"` // 兼容旧字段，后续不再使用
	SortOrder     *int            `json:"sort_order"`
	Touched       map[string]bool `json:"-"`
}

type DailyTaskInput struct {
	StageGoalID      *uint           `json:"stage_goal_id"`
	StageItemID      *uint           `json:"stage_item_id"`
	WeeklyTaskID     *uint           `json:"weekly_task_id"`
	Title            string          `json:"title"`
	Note             string          `json:"note"`
	Date             *string         `json:"date"`
	PlanDate         *string         `json:"plan_date"` // 兼容旧前端字段
	Deadline         *string         `json:"deadline"`
	DueDate          *string         `json:"due_date"` // 兼容旧前端字段
	Status           string          `json:"status"`
	Priority         string          `json:"priority"`
	EstimatedMinutes int             `json:"estimated_minutes"`
	Done             *bool           `json:"done"`
	SortOrder        *int            `json:"sort_order"`
	Touched          map[string]bool `json:"-"`
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

type StageItemView struct {
	models.StageItem
	WeeklyTotal     int `json:"weekly_total"`
	WeeklyDone      int `json:"weekly_done"`
	ProgressPercent int `json:"progress_percent"`
}

type WeeklyTaskMaterializeInput struct {
	Mode    string   `json:"mode"`
	Date    *string  `json:"date"`
	Dates   []string `json:"dates"`
	Replace bool     `json:"replace"`
}

type WeeklyTaskMaterializeResult struct {
	WeeklyTask   models.WeeklyTask  `json:"weekly_task"`
	DailyTasks   []models.DailyTask `json:"daily_tasks"`
	CreatedCount int                `json:"created_count"`
	UpdatedCount int                `json:"updated_count"`
	DeletedCount int                `json:"deleted_count"`
}

type StageGoalView struct {
	models.StageGoal
	ItemTotal       int              `json:"item_total"`
	ItemDone        int              `json:"item_done"`
	WeeklyTotal     int              `json:"weekly_total"`
	WeeklyDone      int              `json:"weekly_done"`
	StageItems      []StageItemView  `json:"stage_items"`
	CurrentWeekly   []WeeklyTaskView `json:"current_weekly_tasks"`
	ProgressPercent int              `json:"progress_percent"`
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
		if !task.Done && task.Deadline != nil && task.Deadline.Before(day) {
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
	taskDateText, _ := firstProvidedString(input.Date, input.PlanDate)
	taskDate, err := parseOptionalLocalDay(taskDateText)
	if err != nil {
		return nil, err
	}
	if taskDate == nil && defaultToday {
		day, _ := parseLocalDay("")
		taskDate = &day
	}
	deadlineText, _ := firstProvidedString(input.Deadline, input.DueDate)
	deadline, err := parseOptionalLocalDay(deadlineText)
	if err != nil {
		return nil, err
	}
	stageGoalID, err := s.resolveStageGoalForItem(userID, input.StageGoalID, input.StageItemID)
	if err != nil {
		return nil, err
	}

	task := models.DailyTask{
		BaseModel:        models.BaseModel{UserID: userID},
		StageGoalID:      stageGoalID,
		StageItemID:      input.StageItemID,
		WeeklyTaskID:     input.WeeklyTaskID,
		Title:            title,
		Note:             strings.TrimSpace(input.Note),
		Date:             taskDate,
		Deadline:         deadline,
		Status:           normalizeStatus(input.Status, "todo"),
		Priority:         normalizePriority(input.Priority),
		EstimatedMinutes: input.EstimatedMinutes,
		Origin:           "manual",
	}
	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}
	if input.Done != nil {
		s.applyDoneState(&task, *input.Done)
	} else if task.Status == "done" {
		s.applyDoneState(&task, true)
	}

	return &task, s.db.Create(&task).Error
}

func (s *DailyTaskService) Update(userID, id uint, input DailyTaskInput) (*models.DailyTask, error) {
	task, err := s.getDaily(userID, id)
	if err != nil {
		return nil, err
	}

	if fieldTouched(input.Touched, "title") {
		title := strings.TrimSpace(input.Title)
		if title == "" {
			return nil, errors.New("title is required")
		}
		task.Title = title
	}
	if fieldTouched(input.Touched, "note") {
		task.Note = strings.TrimSpace(input.Note)
	}
	stageGoalTouched := fieldTouched(input.Touched, "stage_goal_id")
	stageItemTouched := fieldTouched(input.Touched, "stage_item_id")
	if stageGoalTouched || stageItemTouched {
		stageGoalID := task.StageGoalID
		stageItemID := task.StageItemID
		if stageGoalTouched {
			stageGoalID = input.StageGoalID
		}
		if stageItemTouched {
			stageItemID = input.StageItemID
		}
		resolvedStageGoalID, err := s.resolveStageGoalForItem(userID, stageGoalID, stageItemID)
		if err != nil {
			return nil, err
		}
		task.StageGoalID = resolvedStageGoalID
		task.StageItemID = stageItemID
	}
	if fieldTouched(input.Touched, "weekly_task_id") {
		task.WeeklyTaskID = input.WeeklyTaskID
	}
	if fieldTouched(input.Touched, "date", "plan_date") {
		taskDateText, _ := firstProvidedString(input.Date, input.PlanDate)
		taskDate, err := parseOptionalLocalDay(taskDateText)
		if err != nil {
			return nil, err
		}
		task.Date = taskDate
	}
	if fieldTouched(input.Touched, "deadline", "due_date") {
		deadlineText, _ := firstProvidedString(input.Deadline, input.DueDate)
		deadline, err := parseOptionalLocalDay(deadlineText)
		if err != nil {
			return nil, err
		}
		task.Deadline = deadline
	}
	if fieldTouched(input.Touched, "status") {
		task.Status = normalizeStatus(input.Status, "todo")
		task.Done = task.Status == "done"
		if task.Done && task.DoneAt == nil {
			now := time.Now()
			task.DoneAt = &now
		}
		if !task.Done {
			task.DoneAt = nil
		}
	}
	if fieldTouched(input.Touched, "priority") {
		task.Priority = normalizePriority(input.Priority)
	}
	if fieldTouched(input.Touched, "estimated_minutes") {
		task.EstimatedMinutes = input.EstimatedMinutes
	}
	if fieldTouched(input.Touched, "done") && input.Done != nil {
		s.applyDoneState(task, *input.Done)
	}
	if fieldTouched(input.Touched, "sort_order") && input.SortOrder != nil {
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

func (s *DailyTaskService) ListDailyTasks(userID uint, date string, unscheduled bool, weeklyTaskID string, deadline string) ([]models.DailyTask, error) {
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
	} else if deadline != "" {
		day, err := parseLocalDay(deadline)
		if err != nil {
			return nil, err
		}
		query = query.Where("due_date >= ? AND due_date < ?", day, day.AddDate(0, 0, 1))
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
	currentWeekStart, _ := weekStartDate("")
	currentWeekKey := currentWeekStart.Format("2006-01-02")
	for _, goal := range goals {
		view := StageGoalView{StageGoal: goal}
		stageGoalIDText := strconv.FormatUint(uint64(goal.ID), 10)
		items, err := s.ListStageItems(userID, stageGoalIDText)
		if err != nil {
			return nil, err
		}
		view.StageItems = items
		view.ItemTotal = len(items)
		for _, item := range items {
			if item.Status == "done" {
				view.ItemDone++
			}
		}
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
		currentWeekly, err := s.ListWeeklyTasks(userID, currentWeekKey, stageGoalIDText, "")
		if err != nil {
			return nil, err
		}
		for _, task := range currentWeekly {
			if task.Status == "doing" || task.Status == "todo" {
				view.CurrentWeekly = append(view.CurrentWeekly, task)
			}
			if len(view.CurrentWeekly) >= 3 {
				break
			}
		}
		view.ProgressPercent = percent(view.WeeklyDone, view.WeeklyTotal)
		views = append(views, view)
	}
	return views, nil
}

func (s *DailyTaskService) CreateStageGoal(userID uint, input StageGoalInput) (*models.StageGoal, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	start, err := parseOptionalLocalDay(stringValue(input.StartDate))
	if err != nil {
		return nil, err
	}
	end, err := parseOptionalLocalDay(stringValue(input.EndDate))
	if err != nil {
		return nil, err
	}
	goal := models.StageGoal{
		BaseModel:    models.BaseModel{UserID: userID},
		Title:        title,
		Note:         strings.TrimSpace(input.Note),
		StartDate:    start,
		EndDate:      end,
		TargetText:   strings.TrimSpace(firstNonEmpty(input.TargetText, input.TargetMetric)),
		Status:       normalizeStatus(input.Status, "active"),
		ProgressMode: normalizeProgressMode(input.ProgressMode),
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
	if fieldTouched(input.Touched, "title") {
		title := strings.TrimSpace(input.Title)
		if title == "" {
			return nil, errors.New("title is required")
		}
		goal.Title = title
	}
	if fieldTouched(input.Touched, "note") {
		goal.Note = strings.TrimSpace(input.Note)
	}
	if fieldTouched(input.Touched, "target_text", "target_metric") {
		goal.TargetText = strings.TrimSpace(firstNonEmpty(input.TargetText, input.TargetMetric))
	}
	if fieldTouched(input.Touched, "start_date") {
		start, err := parseOptionalLocalDay(stringValue(input.StartDate))
		if err != nil {
			return nil, err
		}
		goal.StartDate = start
	}
	if fieldTouched(input.Touched, "end_date") {
		end, err := parseOptionalLocalDay(stringValue(input.EndDate))
		if err != nil {
			return nil, err
		}
		goal.EndDate = end
	}
	if fieldTouched(input.Touched, "status") {
		goal.Status = normalizeStatus(input.Status, "active")
	}
	if fieldTouched(input.Touched, "progress_mode") {
		goal.ProgressMode = normalizeProgressMode(input.ProgressMode)
	}
	if fieldTouched(input.Touched, "sort_order") && input.SortOrder != nil {
		goal.SortOrder = *input.SortOrder
	}
	return &goal, s.db.Save(&goal).Error
}

func (s *DailyTaskService) DeleteStageGoal(userID, id uint) error {
	return deleteByUserID[models.StageGoal](s.db, userID, id)
}

func (s *DailyTaskService) ListStageItems(userID uint, stageGoalID string) ([]StageItemView, error) {
	query := s.db.Where("user_id = ?", userID)
	if stageGoalID != "" {
		id, err := strconv.ParseUint(stageGoalID, 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("stage_goal_id = ?", uint(id))
	}

	var items []models.StageItem
	if err := query.Order("status asc, sort_order asc, end_date asc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	views := make([]StageItemView, 0, len(items))
	for _, item := range items {
		view := StageItemView{StageItem: item}
		var weekly []models.WeeklyTask
		if err := s.db.Where("user_id = ? AND stage_item_id = ?", userID, item.ID).Find(&weekly).Error; err != nil {
			return nil, err
		}
		view.WeeklyTotal = len(weekly)
		for _, task := range weekly {
			if task.Status == "done" {
				view.WeeklyDone++
			}
		}
		view.ProgressPercent = percent(view.WeeklyDone, view.WeeklyTotal)
		views = append(views, view)
	}
	return views, nil
}

func (s *DailyTaskService) CreateStageItem(userID uint, input StageItemInput) (*models.StageItem, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	if input.StageGoalID == 0 {
		return nil, errors.New("stage_goal_id is required")
	}
	if err := s.validateStageGoal(userID, input.StageGoalID); err != nil {
		return nil, err
	}
	start, err := parseOptionalLocalDay(stringValue(input.StartDate))
	if err != nil {
		return nil, err
	}
	end, err := parseOptionalLocalDay(stringValue(input.EndDate))
	if err != nil {
		return nil, err
	}
	item := models.StageItem{
		BaseModel:    models.BaseModel{UserID: userID},
		StageGoalID:  input.StageGoalID,
		Title:        title,
		Note:         strings.TrimSpace(input.Note),
		StartDate:    start,
		EndDate:      end,
		TargetText:   strings.TrimSpace(firstNonEmpty(input.TargetText, input.TargetMetric)),
		Status:       normalizeStatus(input.Status, "active"),
		ProgressMode: normalizeProgressMode(input.ProgressMode),
	}
	if input.SortOrder != nil {
		item.SortOrder = *input.SortOrder
	}
	return &item, s.db.Create(&item).Error
}

func (s *DailyTaskService) UpdateStageItem(userID, id uint, input StageItemInput) (*models.StageItem, error) {
	var item models.StageItem
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&item).Error; err != nil {
		return nil, err
	}
	originalStageGoalID := item.StageGoalID
	if fieldTouched(input.Touched, "stage_goal_id") {
		if input.StageGoalID == 0 {
			return nil, errors.New("stage_goal_id is required")
		}
		if err := s.validateStageGoal(userID, input.StageGoalID); err != nil {
			return nil, err
		}
		item.StageGoalID = input.StageGoalID
	}
	if fieldTouched(input.Touched, "title") {
		title := strings.TrimSpace(input.Title)
		if title == "" {
			return nil, errors.New("title is required")
		}
		item.Title = title
	}
	if fieldTouched(input.Touched, "note") {
		item.Note = strings.TrimSpace(input.Note)
	}
	if fieldTouched(input.Touched, "target_text", "target_metric") {
		item.TargetText = strings.TrimSpace(firstNonEmpty(input.TargetText, input.TargetMetric))
	}
	if fieldTouched(input.Touched, "start_date") {
		start, err := parseOptionalLocalDay(stringValue(input.StartDate))
		if err != nil {
			return nil, err
		}
		item.StartDate = start
	}
	if fieldTouched(input.Touched, "end_date") {
		end, err := parseOptionalLocalDay(stringValue(input.EndDate))
		if err != nil {
			return nil, err
		}
		item.EndDate = end
	}
	if fieldTouched(input.Touched, "status") {
		item.Status = normalizeStatus(input.Status, "active")
	}
	if fieldTouched(input.Touched, "progress_mode") {
		item.ProgressMode = normalizeProgressMode(input.ProgressMode)
	}
	if fieldTouched(input.Touched, "sort_order") && input.SortOrder != nil {
		item.SortOrder = *input.SortOrder
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		if item.StageGoalID == originalStageGoalID {
			return nil
		}
		updates := map[string]any{"stage_goal_id": item.StageGoalID}
		if err := tx.Model(&models.WeeklyTask{}).Where("user_id = ? AND stage_item_id = ?", userID, item.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DailyTask{}).Where("user_id = ? AND stage_item_id = ?", userID, item.ID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	return &item, err
}

func (s *DailyTaskService) DeleteStageItem(userID, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.StageItem
		if err := tx.Where("user_id = ? AND id = ?", userID, id).First(&item).Error; err != nil {
			return err
		}
		clearItem := map[string]any{"stage_item_id": nil}
		if err := tx.Model(&models.WeeklyTask{}).Where("user_id = ? AND stage_item_id = ?", userID, id).Updates(clearItem).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DailyTask{}).Where("user_id = ? AND stage_item_id = ?", userID, id).Updates(clearItem).Error; err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND id = ?", userID, id).Delete(&models.StageItem{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *DailyTaskService) ListWeeklyTasks(userID uint, weekStart string, stageGoalID string, stageItemID string) ([]WeeklyTaskView, error) {
	query := s.db.Where("user_id = ?", userID)
	if weekStart != "" {
		start, err := weekStartDate(weekStart)
		if err != nil {
			return nil, err
		}
		end := start.AddDate(0, 0, 6)
		query = query.Where(
			"((task_kind = ? AND COALESCE(start_date, week_start) <= ? AND COALESCE(end_date, week_end, start_date, week_start) >= ?) OR ((task_kind IS NULL OR task_kind = '' OR task_kind <> ?) AND week_start = ?))",
			"special_project",
			end,
			start,
			"special_project",
			start,
		)
	}
	if stageGoalID != "" {
		id, err := strconv.ParseUint(stageGoalID, 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("stage_goal_id = ?", uint(id))
	}
	if stageItemID != "" {
		id, err := strconv.ParseUint(stageItemID, 10, 64)
		if err != nil {
			return nil, err
		}
		query = query.Where("stage_item_id = ?", uint(id))
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
		if task.Status == "done" {
			view.ProgressPercent = 100
		} else {
			view.ProgressPercent = percent(view.DailyDone, view.DailyTotal)
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *DailyTaskService) CreateWeeklyTask(userID uint, input WeeklyTaskInput) (*models.WeeklyTask, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	stageGoalID, err := s.resolveStageGoalForItem(userID, input.StageGoalID, input.StageItemID)
	if err != nil {
		return nil, err
	}
	taskKind := normalizeTaskKind(input.TaskKind)
	window, err := weeklyWindowForCreate(taskKind, input)
	if err != nil {
		return nil, err
	}
	deadlineText, _ := firstProvidedString(input.Deadline, input.DueDate)
	deadline, err := parseOptionalLocalDay(deadlineText)
	if err != nil {
		return nil, err
	}
	task := models.WeeklyTask{
		BaseModel:   models.BaseModel{UserID: userID},
		StageGoalID: stageGoalID,
		StageItemID: input.StageItemID,
		TaskKind:    taskKind,
		Title:       title,
		Note:        strings.TrimSpace(input.Note),
		WeekStart:   window.WeekStart,
		WeekEnd:     window.WeekEnd,
		StartDate:   window.StartDate,
		EndDate:     window.EndDate,
		Deadline:    deadline,
		Priority:    normalizePriority(input.Priority),
		Status:      normalizeStatus(input.Status, "todo"),
		ExecuteMode: normalizeExecuteMode(input.ExecuteMode),
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
	if fieldTouched(input.Touched, "task_kind") {
		task.TaskKind = normalizeTaskKind(input.TaskKind)
	} else {
		task.TaskKind = normalizeTaskKind(task.TaskKind)
	}
	if fieldTouched(input.Touched, "title") {
		title := strings.TrimSpace(input.Title)
		if title == "" {
			return nil, errors.New("title is required")
		}
		task.Title = title
	}
	if fieldTouched(input.Touched, "note") {
		task.Note = strings.TrimSpace(input.Note)
	}
	stageGoalTouched := fieldTouched(input.Touched, "stage_goal_id")
	stageItemTouched := fieldTouched(input.Touched, "stage_item_id")
	if stageGoalTouched || stageItemTouched {
		stageGoalID := task.StageGoalID
		stageItemID := task.StageItemID
		if stageGoalTouched {
			stageGoalID = input.StageGoalID
		}
		if stageItemTouched {
			stageItemID = input.StageItemID
		}
		resolvedStageGoalID, err := s.resolveStageGoalForItem(userID, stageGoalID, stageItemID)
		if err != nil {
			return nil, err
		}
		task.StageGoalID = resolvedStageGoalID
		task.StageItemID = stageItemID
	}
	if err := applyWeeklyWindowUpdate(&task, input); err != nil {
		return nil, err
	}
	if fieldTouched(input.Touched, "deadline", "due_date") {
		deadlineText, _ := firstProvidedString(input.Deadline, input.DueDate)
		deadline, err := parseOptionalLocalDay(deadlineText)
		if err != nil {
			return nil, err
		}
		task.Deadline = deadline
	}
	if fieldTouched(input.Touched, "priority") {
		task.Priority = normalizePriority(input.Priority)
	}
	if fieldTouched(input.Touched, "status") {
		task.Status = normalizeStatus(input.Status, "todo")
	}
	if fieldTouched(input.Touched, "execute_mode") {
		task.ExecuteMode = normalizeExecuteMode(input.ExecuteMode)
	}
	if fieldTouched(input.Touched, "sort_order") && input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}
	return &task, s.db.Save(&task).Error
}

func (s *DailyTaskService) MaterializeWeeklyTask(userID, id uint, input WeeklyTaskMaterializeInput) (*WeeklyTaskMaterializeResult, error) {
	dates, err := materializeDates(input.Mode, input)
	if err != nil {
		return nil, err
	}

	result := &WeeklyTaskMaterializeResult{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var task models.WeeklyTask
		if err := tx.Where("user_id = ? AND id = ?", userID, id).First(&task).Error; err != nil {
			return err
		}
		result.WeeklyTask = task
		if len(dates) == 0 {
			return nil
		}

		rangeStart, rangeEnd := materializeWindow(task)
		targetDates := make(map[string]bool, len(dates))
		for _, dateText := range dates {
			day, err := parseLocalDay(dateText)
			if err != nil {
				return err
			}
			if day.Before(rangeStart) || day.After(rangeEnd) {
				return errors.New("materialize date must be within task range")
			}
			targetDates[day.Format("2006-01-02")] = true
		}
		if input.Replace {
			deleted, err := s.deleteStaleMaterializedDaily(tx, userID, task.ID, targetDates)
			if err != nil {
				return err
			}
			result.DeletedCount = deleted
		}
		for _, dateText := range dates {
			day, err := parseLocalDay(dateText)
			if err != nil {
				return err
			}
			daily, created, err := s.upsertDailyFromWeekly(tx, userID, task, day)
			if err != nil {
				return err
			}
			if created {
				result.CreatedCount++
			} else {
				result.UpdatedCount++
			}
			result.DailyTasks = append(result.DailyTasks, *daily)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *DailyTaskService) deleteStaleMaterializedDaily(tx *gorm.DB, userID uint, weeklyTaskID uint, targetDates map[string]bool) (int, error) {
	var tasks []models.DailyTask
	if err := tx.Where(
		"user_id = ? AND weekly_task_id = ? AND done = ? AND origin = ?",
		userID,
		weeklyTaskID,
		false,
		"weekly_materialized",
	).Find(&tasks).Error; err != nil {
		return 0, err
	}
	deleted := 0
	for _, task := range tasks {
		if task.Date != nil && targetDates[task.Date.Format("2006-01-02")] {
			continue
		}
		result := tx.Where("user_id = ? AND id = ?", userID, task.ID).Delete(&models.DailyTask{})
		if result.Error != nil {
			return deleted, result.Error
		}
		deleted += int(result.RowsAffected)
	}
	return deleted, nil
}

func (s *DailyTaskService) upsertDailyFromWeekly(tx *gorm.DB, userID uint, weekly models.WeeklyTask, day time.Time) (*models.DailyTask, bool, error) {
	var existing models.DailyTask
	err := tx.Where(
		"user_id = ? AND weekly_task_id = ? AND plan_date >= ? AND plan_date < ?",
		userID,
		weekly.ID,
		day,
		day.AddDate(0, 0, 1),
	).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		task := models.DailyTask{
			BaseModel:    models.BaseModel{UserID: userID},
			WeeklyTaskID: &weekly.ID,
			StageGoalID:  weekly.StageGoalID,
			StageItemID:  weekly.StageItemID,
			Title:        weekly.Title,
			Note:         strings.TrimSpace(weekly.Note),
			Date:         &day,
			Deadline:     weekly.Deadline,
			Status:       "todo",
			Priority:     normalizePriority(weekly.Priority),
			Origin:       "weekly_materialized",
		}
		if err := tx.Create(&task).Error; err != nil {
			return nil, false, err
		}
		return &task, true, nil
	}

	existing.StageGoalID = weekly.StageGoalID
	existing.StageItemID = weekly.StageItemID
	existing.Title = weekly.Title
	existing.Note = strings.TrimSpace(weekly.Note)
	existing.Deadline = weekly.Deadline
	existing.Priority = normalizePriority(weekly.Priority)
	if existing.Origin == "" {
		existing.Origin = "weekly_materialized"
	}
	if !existing.Done && existing.Status == "" {
		existing.Status = "todo"
	}
	if err := tx.Save(&existing).Error; err != nil {
		return nil, false, err
	}
	return &existing, false, nil
}

func (s *DailyTaskService) DeleteWeeklyTask(userID, id uint) error {
	return deleteByUserID[models.WeeklyTask](s.db, userID, id)
}

func (s *DailyTaskService) applyDoneState(task *models.DailyTask, done bool) {
	task.Done = done
	if done {
		now := time.Now()
		task.DoneAt = &now
		task.Status = "done"
	} else {
		task.DoneAt = nil
		if task.Status == "done" {
			task.Status = "todo"
		}
	}
}

func (s *DailyTaskService) getDaily(userID, id uint) (*models.DailyTask, error) {
	var task models.DailyTask
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *DailyTaskService) validateStageGoal(userID, id uint) error {
	if id == 0 {
		return errors.New("stage_goal_id is required")
	}
	var goal models.StageGoal
	return s.db.Where("user_id = ? AND id = ?", userID, id).First(&goal).Error
}

func (s *DailyTaskService) getStageItem(userID, id uint) (*models.StageItem, error) {
	if id == 0 {
		return nil, errors.New("stage_item_id is required")
	}
	var item models.StageItem
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *DailyTaskService) resolveStageGoalForItem(userID uint, stageGoalID *uint, stageItemID *uint) (*uint, error) {
	if stageItemID == nil {
		if stageGoalID != nil {
			if err := s.validateStageGoal(userID, *stageGoalID); err != nil {
				return nil, err
			}
		}
		return stageGoalID, nil
	}

	item, err := s.getStageItem(userID, *stageItemID)
	if err != nil {
		return nil, err
	}
	if stageGoalID != nil && *stageGoalID != item.StageGoalID {
		return nil, errors.New("stage_item_id does not belong to stage_goal_id")
	}
	resolved := item.StageGoalID
	return &resolved, nil
}

type weeklyTaskWindow struct {
	WeekStart time.Time
	WeekEnd   *time.Time
	StartDate *time.Time
	EndDate   *time.Time
}

func weeklyWindowForCreate(taskKind string, input WeeklyTaskInput) (weeklyTaskWindow, error) {
	if taskKind == "special_project" {
		startText := stringValue(input.StartDate)
		if startText == "" {
			startText = strings.TrimSpace(input.WeekStart)
		}
		start, err := parseLocalDay(startText)
		if err != nil {
			return weeklyTaskWindow{}, err
		}
		endText := stringValue(input.EndDate)
		if endText == "" {
			endText = stringValue(input.WeekEnd)
		}
		end, err := parseOptionalLocalDay(endText)
		if err != nil {
			return weeklyTaskWindow{}, err
		}
		if end == nil {
			end = copyTime(start)
		}
		if end.Before(start) {
			return weeklyTaskWindow{}, errors.New("end_date cannot be before start_date")
		}
		weekStart, err := weekStartDate(start.Format("2006-01-02"))
		if err != nil {
			return weeklyTaskWindow{}, err
		}
		return weeklyTaskWindow{
			WeekStart: weekStart,
			WeekEnd:   copyOptionalTime(end),
			StartDate: copyTime(start),
			EndDate:   copyOptionalTime(end),
		}, nil
	}

	weekStart, err := weekStartDate(input.WeekStart)
	if err != nil {
		return weeklyTaskWindow{}, err
	}
	weekEnd, err := parseOptionalLocalDay(stringValue(input.WeekEnd))
	if err != nil {
		return weeklyTaskWindow{}, err
	}
	if weekEnd == nil {
		weekEnd = copyTime(weekStart.AddDate(0, 0, 6))
	}
	return weeklyTaskWindow{
		WeekStart: weekStart,
		WeekEnd:   copyOptionalTime(weekEnd),
		StartDate: copyTime(weekStart),
		EndDate:   copyOptionalTime(weekEnd),
	}, nil
}

func applyWeeklyWindowUpdate(task *models.WeeklyTask, input WeeklyTaskInput) error {
	if normalizeTaskKind(task.TaskKind) == "special_project" {
		return applySpecialProjectWindowUpdate(task, input)
	}
	return applyStandardWeekWindowUpdate(task, input)
}

func applyStandardWeekWindowUpdate(task *models.WeeklyTask, input WeeklyTaskInput) error {
	if fieldTouched(input.Touched, "week_start") {
		if strings.TrimSpace(input.WeekStart) == "" {
			return errors.New("week_start is required")
		}
		weekStart, err := weekStartDate(input.WeekStart)
		if err != nil {
			return err
		}
		task.WeekStart = weekStart
	}
	if fieldTouched(input.Touched, "week_end") {
		weekEnd, err := parseOptionalLocalDay(stringValue(input.WeekEnd))
		if err != nil {
			return err
		}
		task.WeekEnd = weekEnd
	} else if fieldTouched(input.Touched, "task_kind") {
		task.WeekEnd = copyTime(task.WeekStart.AddDate(0, 0, 6))
	}
	task.StartDate = copyTime(task.WeekStart)
	task.EndDate = copyOptionalTime(task.WeekEnd)
	return nil
}

func applySpecialProjectWindowUpdate(task *models.WeeklyTask, input WeeklyTaskInput) error {
	start := task.WeekStart
	if task.StartDate != nil {
		start = *task.StartDate
	}
	if fieldTouched(input.Touched, "start_date") {
		parsed, err := parseOptionalLocalDay(stringValue(input.StartDate))
		if err != nil {
			return err
		}
		if parsed != nil {
			start = *parsed
		}
	} else if fieldTouched(input.Touched, "week_start") && strings.TrimSpace(input.WeekStart) != "" {
		parsed, err := parseLocalDay(input.WeekStart)
		if err != nil {
			return err
		}
		start = parsed
	}

	end := start
	if task.EndDate != nil {
		end = *task.EndDate
	} else if task.WeekEnd != nil {
		end = *task.WeekEnd
	}
	if fieldTouched(input.Touched, "end_date") {
		parsed, err := parseOptionalLocalDay(stringValue(input.EndDate))
		if err != nil {
			return err
		}
		if parsed != nil {
			end = *parsed
		} else {
			end = start
		}
	} else if fieldTouched(input.Touched, "week_end") {
		parsed, err := parseOptionalLocalDay(stringValue(input.WeekEnd))
		if err != nil {
			return err
		}
		if parsed != nil {
			end = *parsed
		} else {
			end = start
		}
	}
	if end.Before(start) {
		return errors.New("end_date cannot be before start_date")
	}

	weekStart, err := weekStartDate(start.Format("2006-01-02"))
	if err != nil {
		return err
	}
	task.WeekStart = weekStart
	task.WeekEnd = copyTime(end)
	task.StartDate = copyTime(start)
	task.EndDate = copyTime(end)
	return nil
}

func materializeWindow(task models.WeeklyTask) (time.Time, time.Time) {
	if normalizeTaskKind(task.TaskKind) == "special_project" {
		start := task.WeekStart
		if task.StartDate != nil {
			start = *task.StartDate
		}
		end := start
		if task.EndDate != nil {
			end = *task.EndDate
		} else if task.WeekEnd != nil {
			end = *task.WeekEnd
		}
		return start, end
	}

	end := task.WeekStart.AddDate(0, 0, 6)
	if task.WeekEnd != nil {
		end = *task.WeekEnd
	}
	return task.WeekStart, end
}

func copyTime(value time.Time) *time.Time {
	copied := value
	return &copied
}

func copyOptionalTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	return copyTime(*value)
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

func normalizeTaskKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "special_project":
		return "special_project"
	default:
		return "standard_week"
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

func normalizeExecuteMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "scheduled_day", "split_to_days", "ddl_only":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "weekly_todo"
	}
}

func normalizeProgressMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "manual", "weekly":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "weekly"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstProvidedString(values ...*string) (string, bool) {
	for _, value := range values {
		if value != nil {
			return strings.TrimSpace(*value), true
		}
	}
	return "", false
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func fieldTouched(fields map[string]bool, names ...string) bool {
	if fields == nil {
		return true
	}
	for _, name := range names {
		if fields[name] {
			return true
		}
	}
	return false
}

func materializeDates(mode string, input WeeklyTaskMaterializeInput) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "weekly_todo", "ddl_only":
		return nil, nil
	case "scheduled_day":
		if input.Date == nil || strings.TrimSpace(*input.Date) == "" {
			return nil, errors.New("date is required for scheduled_day materialization")
		}
		return normalizeDateKeys([]string{*input.Date})
	case "split_to_days":
		if len(input.Dates) == 0 {
			return nil, errors.New("dates are required for split_to_days materialization")
		}
		return normalizeDateKeys(input.Dates)
	default:
		return nil, errors.New("unsupported materialize mode")
	}
}

func normalizeDateKeys(values []string) ([]string, error) {
	seen := make(map[string]bool, len(values))
	dates := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, errors.New("materialize date is required")
		}
		day, err := parseLocalDay(value)
		if err != nil {
			return nil, err
		}
		key := day.Format("2006-01-02")
		if seen[key] {
			continue
		}
		seen[key] = true
		dates = append(dates, key)
	}
	if len(dates) > 14 {
		return nil, errors.New("materialize dates cannot exceed 14 days")
	}
	return dates, nil
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
