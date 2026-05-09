package services

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type StudyLogStats struct {
	TotalMinutes  int    `json:"total_minutes"`
	PomodoroCount int64  `json:"pomodoro_count"`
	Interruptions int    `json:"interruptions"`
	MainSubject   string `json:"main_subject"`
}

type CalendarEvent struct {
	ID     string    `json:"id"`
	Type   string    `json:"type"`
	Title  string    `json:"title"`
	Date   time.Time `json:"date"`
	Status string    `json:"status"`
}

type StudyService struct {
	db *gorm.DB
}

func NewStudyService(db *gorm.DB) *StudyService {
	return &StudyService{db: db}
}

func (s *StudyService) ListLogs(userID uint, date string, scope string) ([]models.StudyLog, error) {
	var logs []models.StudyLog
	query := s.db.Where("user_id = ?", userID)
	if date != "" {
		start, end, err := logRange(date, scope)
		if err == nil {
			query = query.Where("start_time >= ? AND start_time < ?", start, end)
		}
	}
	err := query.Order("start_time desc").Find(&logs).Error
	return logs, err
}

func (s *StudyService) LogStats(userID uint, date string, scope string) (*StudyLogStats, error) {
	logs, err := s.ListLogs(userID, date, scope)
	if err != nil {
		return nil, err
	}

	stats := &StudyLogStats{}
	subjectMinutes := map[string]int{}
	for _, log := range logs {
		stats.TotalMinutes += log.DurationMin
		if log.Source == "pomodoro" {
			stats.PomodoroCount++
		}
		if log.Subject != "" {
			subjectMinutes[log.Subject] += log.DurationMin
		}
	}

	for subject, minutes := range subjectMinutes {
		if stats.MainSubject == "" || minutes > subjectMinutes[stats.MainSubject] {
			stats.MainSubject = subject
		}
	}

	return stats, nil
}

func (s *StudyService) CreateLog(log *models.StudyLog) error {
	return s.db.Create(log).Error
}

func (s *StudyService) ListPlans(userID uint, planType string) ([]models.StudyPlan, error) {
	var plans []models.StudyPlan
	query := s.db.Where("user_id = ?", userID)
	if planType != "" {
		query = query.Where("plan_type = ?", planType)
	}
	err := query.Order("due_date asc nulls last, created_at desc").Find(&plans).Error
	return plans, err
}

func (s *StudyService) CreatePlan(plan *models.StudyPlan) error {
	return s.db.Create(plan).Error
}

func (s *StudyService) UpdatePlan(userID uint, id uint, updates *models.StudyPlan) (*models.StudyPlan, error) {
	var plan models.StudyPlan
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&plan).Error; err != nil {
		return nil, err
	}

	plan.Title = updates.Title
	plan.PlanType = updates.PlanType
	plan.Subject = updates.Subject
	plan.QuestionType = updates.QuestionType
	plan.TargetMin = updates.TargetMin
	plan.TargetCount = updates.TargetCount
	plan.StartDate = updates.StartDate
	plan.DueDate = updates.DueDate
	plan.Priority = updates.Priority
	plan.Status = updates.Status
	plan.Note = updates.Note

	return &plan, s.db.Save(&plan).Error
}

func (s *StudyService) DeletePlan(userID uint, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.StudyPlan{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *StudyService) CompletePlan(userID uint, id uint) (*models.StudyPlan, error) {
	var plan models.StudyPlan
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&plan).Error; err != nil {
		return nil, err
	}
	plan.Status = "已完成"
	return &plan, s.db.Save(&plan).Error
}

func (s *StudyService) CalendarEvents(userID uint, month string) ([]CalendarEvent, error) {
	start, end := monthRange(month)
	events := []CalendarEvent{}

	var plans []models.StudyPlan
	if err := s.db.Where("user_id = ? AND due_date >= ? AND due_date < ?", userID, start, end).Find(&plans).Error; err != nil {
		return nil, err
	}
	for _, plan := range plans {
		if plan.DueDate != nil {
			events = append(events, CalendarEvent{ID: idString("plan", plan.ID), Type: "学习计划", Title: plan.Title, Date: *plan.DueDate, Status: plan.Status})
		}
	}

	var logs []models.StudyLog
	if err := s.db.Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, start, end).Find(&logs).Error; err != nil {
		return nil, err
	}
	for _, log := range logs {
		if log.StartTime != nil {
			events = append(events, CalendarEvent{ID: idString("log", log.ID), Type: "番茄钟学习记录", Title: log.Note, Date: *log.StartTime, Status: log.StudyType})
		}
	}

	var mistakes []models.Mistake
	if err := s.db.Where("user_id = ? AND next_review_at >= ? AND next_review_at < ?", userID, start, end).Find(&mistakes).Error; err != nil {
		return nil, err
	}
	for _, mistake := range mistakes {
		if mistake.NextReviewAt != nil {
			events = append(events, CalendarEvent{ID: idString("mistake", mistake.ID), Type: "错题复习", Title: mistake.Title, Date: *mistake.NextReviewAt, Status: mistake.Mastery})
		}
	}

	return events, nil
}

func dayRange(date string) (time.Time, time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return parsed, parsed.AddDate(0, 0, 1), nil
}

func logRange(date string, scope string) (time.Time, time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	switch strings.ToLower(scope) {
	case "week":
		weekday := int(parsed.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location()).AddDate(0, 0, -(weekday - 1))
		return start, start.AddDate(0, 0, 7), nil
	case "month":
		start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, parsed.Location())
		return start, start.AddDate(0, 1, 0), nil
	default:
		return dayRange(date)
	}
}

func monthRange(month string) (time.Time, time.Time) {
	if month == "" {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 1, 0)
	}
	parsed, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		now := time.Now()
		parsed = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	return parsed, parsed.AddDate(0, 1, 0)
}

func idString(prefix string, id uint) string {
	return prefix + "-" + strconvFormatUint(id)
}

func strconvFormatUint(id uint) string {
	return fmt.Sprintf("%d", id)
}
