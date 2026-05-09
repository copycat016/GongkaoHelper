package models

import "time"

type StudyLog struct {
	BaseModel
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	DurationMin  int        `json:"duration_min" gorm:"not null;default:0"`
	StudyType    string     `json:"study_type" gorm:"size:80;index"`
	Subject      string     `json:"subject" gorm:"size:60;index"`
	QuestionType string     `json:"question_type" gorm:"size:120"`
	Source       string     `json:"source" gorm:"size:120"`
	SourceID     uint       `json:"source_id" gorm:"index"`
	Note         string     `json:"note" gorm:"type:text"`
}

type StudyPlan struct {
	BaseModel
	Title        string     `json:"title" gorm:"size:200;not null"`
	PlanType     string     `json:"plan_type" gorm:"size:60;index"`
	Subject      string     `json:"subject" gorm:"size:60;index"`
	QuestionType string     `json:"question_type" gorm:"size:120"`
	TargetMin    int        `json:"target_min" gorm:"not null;default:0"`
	TargetCount  int        `json:"target_count" gorm:"not null;default:0"`
	StartDate    *time.Time `json:"start_date" gorm:"index"`
	DueDate      *time.Time `json:"due_date" gorm:"index"`
	Priority     string     `json:"priority" gorm:"size:40;index"`
	Status       string     `json:"status" gorm:"size:40;index;default:进行中"`
	Note         string     `json:"note" gorm:"type:text"`
}
