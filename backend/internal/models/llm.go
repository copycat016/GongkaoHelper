package models

type LLMProvider struct {
	BaseModel
	Name         string `json:"name" gorm:"size:120;not null;index:idx_llm_provider_user_name,unique"`
	ProviderType string `json:"provider_type" gorm:"size:60;not null;default:openai-compatible"`
	BaseURL      string `json:"base_url" gorm:"size:500;not null"`
	APIKey       string `json:"api_key,omitempty" gorm:"size:1000"`
	APIKeyMasked string `json:"api_key_masked" gorm:"-"`
	HasAPIKey    bool   `json:"has_api_key" gorm:"-"`
	Enabled      bool   `json:"enabled" gorm:"not null;default:true"`
	Note         string `json:"note" gorm:"size:1000"`
}

type LLMModel struct {
	BaseModel
	ProviderID   uint   `json:"provider_id" gorm:"not null;index"`
	Provider     string `json:"provider" gorm:"size:120"`
	Name         string `json:"name" gorm:"size:160;not null"`
	Alias        string `json:"alias" gorm:"size:160"`
	CostLevel    string `json:"cost_level" gorm:"size:20"`
	SpeedLevel   string `json:"speed_level" gorm:"size:20"`
	QualityLevel string `json:"quality_level" gorm:"size:20"`
	Enabled      bool   `json:"enabled" gorm:"not null;default:true"`
	UseFast      bool   `json:"use_fast" gorm:"not null;default:false"`
	UseQuality   bool   `json:"use_quality" gorm:"not null;default:false"`
	UseOCR       bool   `json:"use_ocr" gorm:"not null;default:false"`
	UseQuestion  bool   `json:"use_question" gorm:"not null;default:false"`
	UseEssay     bool   `json:"use_essay" gorm:"not null;default:false"`
	UseSummary   bool   `json:"use_summary" gorm:"not null;default:false"`
	UsePlan      bool   `json:"use_plan" gorm:"not null;default:false"`
}
