package models

type EssayDocument struct {
	BaseModel
	Title         string `json:"title" gorm:"size:200;not null"`
	DocumentRole  string `json:"document_role" gorm:"size:40;not null;default:combined;index"`
	SourceGroup   string `json:"source_group" gorm:"size:160;index"`
	OriginalName  string `json:"original_name" gorm:"size:300"`
	FilePath      string `json:"file_path" gorm:"size:500"`
	Status        string `json:"status" gorm:"size:40;not null;default:uploaded;index"`
	PageCount     int    `json:"page_count" gorm:"not null;default:0"`
	ChunkCount    int    `json:"chunk_count" gorm:"not null;default:0"`
	ClassModelID  uint   `json:"class_model_id" gorm:"index"`
	ReviewModelID uint   `json:"review_model_id" gorm:"index"`
	Note          string `json:"note" gorm:"size:1000"`
}

type EssayChunk struct {
	BaseModel
	DocumentID         uint    `json:"document_id" gorm:"not null;index"`
	PageNo             int     `json:"page_no" gorm:"not null;default:1;index"`
	ChunkIndex         int     `json:"chunk_index" gorm:"not null;index"`
	Content            string  `json:"content" gorm:"type:text;not null"`
	ChunkType          string  `json:"chunk_type" gorm:"size:40;not null;default:unknown;index"`
	Confidence         float64 `json:"confidence" gorm:"not null;default:0"`
	ClassModelID       uint    `json:"class_model_id" gorm:"index"`
	ClassificationNote string  `json:"classification_note" gorm:"size:1000"`
}

type EssayQuestion struct {
	BaseModel
	DocumentID   uint   `json:"document_id" gorm:"not null;index"`
	Title        string `json:"title" gorm:"size:240;not null"`
	QuestionType string `json:"question_type" gorm:"size:80"`
	QuestionText string `json:"question_text" gorm:"type:text;not null"`
	MaxScore     int    `json:"max_score" gorm:"not null;default:100"`
	WordLimit    int    `json:"word_limit" gorm:"not null;default:500"`
	Status       string `json:"status" gorm:"size:40;not null;default:assembled;index"`
}

type EssayQuestionChunk struct {
	BaseModel
	DocumentID   uint   `json:"document_id" gorm:"not null;index"`
	QuestionID   uint   `json:"question_id" gorm:"not null;index"`
	ChunkID      uint   `json:"chunk_id" gorm:"not null;index"`
	RelationType string `json:"relation_type" gorm:"size:40;not null;index"`
}

type EssayReview struct {
	BaseModel
	QuestionID    uint    `json:"question_id" gorm:"not null;index"`
	ReviewModelID uint    `json:"review_model_id" gorm:"index"`
	UserAnswer    string  `json:"user_answer" gorm:"type:text;not null"`
	Score         float64 `json:"score" gorm:"not null;default:0"`
	MaxScore      int     `json:"max_score" gorm:"not null;default:100"`
	ResultJSON    string  `json:"result_json" gorm:"type:text"`
}
