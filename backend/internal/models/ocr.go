package models

type OCRTask struct {
	BaseModel
	Provider       string `json:"provider" gorm:"size:60;index"`
	Engine         string `json:"engine" gorm:"size:80;index"`
	Scene          string `json:"scene" gorm:"size:80;index"`
	Status         string `json:"status" gorm:"size:40;index"`
	FileName       string `json:"file_name" gorm:"size:255"`
	FileSHA256     string `json:"file_sha256" gorm:"size:80;index"`
	MimeType       string `json:"mime_type" gorm:"size:120"`
	SizeBytes      int64  `json:"size_bytes"`
	ExternalTaskID string `json:"external_task_id" gorm:"size:160;index"`
	RawText        string `json:"raw_text" gorm:"type:text"`
	RawResponse    string `json:"raw_response" gorm:"type:text"`
	ErrorMessage   string `json:"error_message" gorm:"type:text"`
}
