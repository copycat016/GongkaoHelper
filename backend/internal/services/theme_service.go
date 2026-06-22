package services

import (
	"gkweb/backend/internal/models"
	"gorm.io/gorm"
)

type ThemeService struct {
	db *gorm.DB
}

func NewThemeService(db *gorm.DB) *ThemeService {
	return &ThemeService{db: db}
}

func (s *ThemeService) Get(userID uint) (*models.ThemeConfig, error) {
	var config models.ThemeConfig
	err := s.db.Where("user_id = ?", userID).Order("id").Limit(1).Find(&config).Error
	if err != nil {
		return nil, err
	}
	if config.ID == 0 {
		return nil, nil
	}
	return &config, nil
}

func (s *ThemeService) Save(userID uint, config *models.ThemeConfig) (*models.ThemeConfig, error) {
	var existing models.ThemeConfig
	err := s.db.Where("user_id = ?", userID).Order("id").Limit(1).Find(&existing).Error
	if err != nil {
		return nil, err
	}
	if existing.ID == 0 {
		config.UserID = userID
		if err := s.db.Create(config).Error; err != nil {
			return nil, err
		}
		return config, nil
	}

	existing.Palette = config.Palette
	existing.BackgroundEnabled = config.BackgroundEnabled
	existing.BackgroundImage = config.BackgroundImage
	existing.Blur = config.Blur
	existing.Brightness = config.Brightness
	existing.MaskOpacity = config.MaskOpacity
	existing.BackgroundSize = config.BackgroundSize
	existing.BackgroundPosition = config.BackgroundPosition
	existing.CardOpacity = config.CardOpacity
	existing.DockImage = config.DockImage

	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}
