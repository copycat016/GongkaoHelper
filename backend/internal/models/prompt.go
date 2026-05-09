package models

type PromptTemplate struct {
	BaseModel
	TaskType       string `json:"task_type" gorm:"size:120;not null;index"`
	Name           string `json:"name" gorm:"size:160;not null"`
	SystemPrompt   string `json:"system_prompt" gorm:"type:text"`
	UserPrompt     string `json:"user_prompt" gorm:"type:text"`
	Variables      string `json:"variables" gorm:"type:text"`
	DefaultModelID *uint  `json:"default_model_id"`
	DefaultModel   string `json:"default_model" gorm:"size:160"`
	Version        string `json:"version" gorm:"size:40;not null;default:v1.0"`
	Enabled        bool   `json:"enabled" gorm:"not null;default:true"`
}
