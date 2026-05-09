package services

import (
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type PromptService struct {
	db *gorm.DB
}

func NewPromptService(db *gorm.DB) *PromptService {
	return &PromptService{db: db}
}

func (s *PromptService) List(userID uint, taskType string) ([]models.PromptTemplate, error) {
	var prompts []models.PromptTemplate
	query := s.db.Where("user_id = ?", userID)
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	err := query.Order("created_at desc").Find(&prompts).Error
	return prompts, err
}

func (s *PromptService) Create(prompt *models.PromptTemplate) error {
	return s.db.Create(prompt).Error
}

func (s *PromptService) Update(userID uint, id uint, updates *models.PromptTemplate) (*models.PromptTemplate, error) {
	var prompt models.PromptTemplate
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&prompt).Error; err != nil {
		return nil, err
	}

	prompt.TaskType = updates.TaskType
	prompt.Name = updates.Name
	prompt.SystemPrompt = updates.SystemPrompt
	prompt.UserPrompt = updates.UserPrompt
	prompt.Variables = updates.Variables
	prompt.DefaultModelID = updates.DefaultModelID
	prompt.DefaultModel = updates.DefaultModel
	prompt.Version = updates.Version
	prompt.Enabled = updates.Enabled

	return &prompt, s.db.Save(&prompt).Error
}

func (s *PromptService) Delete(userID uint, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.PromptTemplate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
