package services

import (
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type MistakeFilters struct {
	Subject      string
	QuestionType string
	ErrorReason  string
	Mastery      string
	Tag          string
}

type MistakeReviewInput struct {
	Mastery      string     `json:"mastery"`
	NextReviewAt *time.Time `json:"next_review_at"`
	Note         string     `json:"note"`
}

type MistakeService struct {
	db *gorm.DB
}

func NewMistakeService(db *gorm.DB) *MistakeService {
	return &MistakeService{db: db}
}

func (s *MistakeService) List(userID uint, filters MistakeFilters) ([]models.Mistake, error) {
	var mistakes []models.Mistake
	query := s.db.Where("user_id = ?", userID)

	if filters.Subject != "" {
		query = query.Where("subject = ?", filters.Subject)
	}
	if filters.QuestionType != "" {
		query = query.Where("question_type = ?", filters.QuestionType)
	}
	if filters.ErrorReason != "" {
		query = query.Where("error_reason = ?", filters.ErrorReason)
	}
	if filters.Mastery != "" {
		query = query.Where("mastery = ?", filters.Mastery)
	}
	if filters.Tag != "" {
		query = query.Where("tags::text ILIKE ?", "%"+filters.Tag+"%")
	}

	err := query.Order("updated_at desc").Find(&mistakes).Error
	return mistakes, err
}

func (s *MistakeService) Create(mistake *models.Mistake) error {
	return s.db.Create(mistake).Error
}

func (s *MistakeService) Get(userID uint, id uint) (*models.Mistake, error) {
	var mistake models.Mistake
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&mistake).Error; err != nil {
		return nil, err
	}
	return &mistake, nil
}

func (s *MistakeService) Update(userID uint, id uint, updates *models.Mistake) (*models.Mistake, error) {
	mistake, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}

	mistake.Subject = updates.Subject
	mistake.QuestionType = updates.QuestionType
	mistake.SubType = updates.SubType
	mistake.Title = updates.Title
	mistake.Stem = updates.Stem
	mistake.Options = updates.Options
	mistake.CorrectAnswer = updates.CorrectAnswer
	mistake.UserAnswer = updates.UserAnswer
	mistake.Analysis = updates.Analysis
	mistake.ErrorReason = updates.ErrorReason
	mistake.Mastery = updates.Mastery
	mistake.NextReviewAt = updates.NextReviewAt
	mistake.Tags = updates.Tags
	mistake.Source = updates.Source
	mistake.Note = updates.Note

	return mistake, s.db.Save(mistake).Error
}

func (s *MistakeService) Delete(userID uint, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Mistake{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *MistakeService) Review(userID uint, id uint, input MistakeReviewInput) (*models.Mistake, error) {
	mistake, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	mistake.ReviewCount++
	mistake.LastReviewAt = &now
	if input.Mastery != "" {
		mistake.Mastery = input.Mastery
	}
	if input.NextReviewAt != nil {
		mistake.NextReviewAt = input.NextReviewAt
	}
	if input.Note != "" {
		mistake.Note = input.Note
	}

	return mistake, s.db.Save(mistake).Error
}
