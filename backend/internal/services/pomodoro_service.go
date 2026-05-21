package services

import (
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type PomodoroStats struct {
	FocusCount   int64 `json:"focus_count"`
	FocusMinutes int   `json:"focus_minutes"`
	BreakCount   int64 `json:"break_count"`
	BreakMinutes int   `json:"break_minutes"`
}

type PomodoroService struct {
	db *gorm.DB
}

func NewPomodoroService(db *gorm.DB) *PomodoroService {
	return &PomodoroService{db: db}
}

func (s *PomodoroService) CreateSession(session *models.PomodoroSession) error {
	if session.CompletedAt.IsZero() {
		session.CompletedAt = time.Now()
	}
	if session.ActualMinutes == 0 {
		session.ActualMinutes = session.PlannedMinutes
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}

		if session.Mode != "focus" {
			return nil
		}

		startedAt := session.StartedAt
		if startedAt == nil {
			start := session.CompletedAt.Add(-time.Duration(session.ActualMinutes) * time.Minute)
			startedAt = &start
		}

		log := models.StudyLog{
			BaseModel:    models.BaseModel{UserID: session.UserID},
			StartTime:    startedAt,
			EndTime:      &session.CompletedAt,
			DurationMin:  session.ActualMinutes,
			StudyType:    session.TaskType,
			Subject:      subjectFromTaskType(session.TaskType),
			QuestionType: session.TaskName,
			Source:       "pomodoro",
			SourceID:     session.ID,
			Note:         session.TaskName,
		}

		return tx.Create(&log).Error
	})
}

func (s *PomodoroService) TodayStats(userID uint) (*PomodoroStats, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)

	var sessions []models.PomodoroSession
	err := s.db.
		Where("user_id = ? AND completed_at >= ? AND completed_at < ?", userID, start, end).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	stats := &PomodoroStats{}
	for _, session := range sessions {
		if session.Mode == "focus" {
			stats.FocusCount++
			stats.FocusMinutes += session.ActualMinutes
			continue
		}
		if session.Mode == "break" {
			stats.BreakCount++
			stats.BreakMinutes += session.ActualMinutes
		}
	}

	return stats, nil
}

func subjectFromTaskType(taskType string) string {
	switch taskType {
	case "行测刷题":
		return "行测"
	case "申论练习":
		return "申论"
	default:
		return ""
	}
}
