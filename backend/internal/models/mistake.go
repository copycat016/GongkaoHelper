package models

import "time"

type Mistake struct {
	BaseModel
	Subject       string     `json:"subject" gorm:"size:40;index"`
	QuestionType  string     `json:"question_type" gorm:"size:120;index"`
	SubType       string     `json:"sub_type" gorm:"size:120"`
	Title         string     `json:"title" gorm:"size:300;not null"`
	Stem          string     `json:"stem" gorm:"type:text"`
	Options       string     `json:"options" gorm:"type:jsonb"`
	CorrectAnswer string     `json:"correct_answer" gorm:"size:200"`
	UserAnswer    string     `json:"user_answer" gorm:"size:200"`
	Analysis      string     `json:"analysis" gorm:"type:text"`
	ErrorReason   string     `json:"error_reason" gorm:"size:80;index"`
	Mastery       string     `json:"mastery" gorm:"size:40;index;default:未掌握"`
	ReviewCount   int        `json:"review_count" gorm:"not null;default:0"`
	NextReviewAt  *time.Time `json:"next_review_at" gorm:"index"`
	LastReviewAt  *time.Time `json:"last_review_at"`
	Tags          string     `json:"tags" gorm:"type:jsonb"`
	Source        string     `json:"source" gorm:"size:120"`
	Note          string     `json:"note" gorm:"type:text"`
}
