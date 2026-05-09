package models

import "time"

type PomodoroSession struct {
	BaseModel
	TaskType       string     `json:"task_type" gorm:"size:80;index"`
	TaskName       string     `json:"task_name" gorm:"size:200"`
	Mode           string     `json:"mode" gorm:"size:20;not null;index"`
	PlannedMinutes int        `json:"planned_minutes" gorm:"not null;default:0"`
	ActualMinutes  int        `json:"actual_minutes" gorm:"not null;default:0"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    time.Time  `json:"completed_at" gorm:"not null;index"`
	Note           string     `json:"note" gorm:"type:text"`
}
