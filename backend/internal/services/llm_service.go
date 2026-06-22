package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type LLMService struct {
	db *gorm.DB
}

type LLMModelCandidate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnedBy string `json:"owned_by,omitempty"`
}

func NewLLMService(db *gorm.DB) *LLMService {
	return &LLMService{db: db}
}

func (s *LLMService) ListProviders(userID uint) ([]models.LLMProvider, error) {
	var providers []models.LLMProvider
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&providers).Error
	return SafeLLMProviders(providers), err
}

func (s *LLMService) CreateProvider(provider *models.LLMProvider) error {
	return s.db.Create(provider).Error
}

func (s *LLMService) UpdateProvider(userID uint, id uint, updates *models.LLMProvider) (*models.LLMProvider, error) {
	var provider models.LLMProvider
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&provider).Error; err != nil {
		return nil, err
	}

	provider.Name = updates.Name
	provider.ProviderType = updates.ProviderType
	provider.BaseURL = updates.BaseURL
	if strings.TrimSpace(updates.APIKey) != "" {
		provider.APIKey = updates.APIKey
	}
	provider.Enabled = updates.Enabled
	provider.Note = updates.Note

	if err := s.db.Save(&provider).Error; err != nil {
		return nil, err
	}
	safe := SafeLLMProvider(provider)
	return &safe, nil
}

func (s *LLMService) DeleteProvider(userID uint, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.LLMProvider{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *LLMService) FetchProviderModels(userID uint, providerID uint) ([]LLMModelCandidate, error) {
	var provider models.LLMProvider
	if err := s.db.Where("user_id = ? AND id = ?", userID, providerID).First(&provider).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(provider.BaseURL) == "" {
		return nil, errors.New("provider base_url is required")
	}

	req, err := http.NewRequest(http.MethodGet, providerModelsURL(provider.BaseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(provider.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(provider.APIKey))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request model list failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("model list request failed: http %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode model list failed: %w", err)
	}

	candidates := normalizeModelCandidates(payload)
	if len(candidates) == 0 {
		return nil, errors.New("no model candidates found in provider response")
	}
	return candidates, nil
}

func (s *LLMService) ListModels(userID uint) ([]models.LLMModel, error) {
	var modelsList []models.LLMModel
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&modelsList).Error
	return modelsList, err
}

func (s *LLMService) CreateModel(model *models.LLMModel) error {
	if model.ProviderID == 0 {
		return errors.New("provider_id is required")
	}
	return s.db.Create(model).Error
}

func (s *LLMService) UpdateModel(userID uint, id uint, updates *models.LLMModel) (*models.LLMModel, error) {
	var model models.LLMModel
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&model).Error; err != nil {
		return nil, err
	}

	model.ProviderID = updates.ProviderID
	model.Provider = updates.Provider
	model.Name = updates.Name
	model.Alias = updates.Alias
	model.CostLevel = updates.CostLevel
	model.SpeedLevel = updates.SpeedLevel
	model.QualityLevel = updates.QualityLevel
	model.Enabled = updates.Enabled
	model.UseFast = updates.UseFast
	model.UseQuality = updates.UseQuality
	model.UseOCR = updates.UseOCR
	model.UseQuestion = updates.UseQuestion
	model.UseEssay = updates.UseEssay
	model.UseSummary = updates.UseSummary
	model.UsePlan = updates.UsePlan

	return &model, s.db.Save(&model).Error
}

func (s *LLMService) DeleteModel(userID uint, id uint) error {
	result := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.LLMModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func providerModelsURL(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(base, "/models") {
		return base
	}
	return base + "/models"
}

func normalizeModelCandidates(payload any) []LLMModelCandidate {
	seen := map[string]bool{}
	candidates := make([]LLMModelCandidate, 0)

	var appendCandidate func(candidate LLMModelCandidate)
	appendCandidate = func(candidate LLMModelCandidate) {
		candidate.ID = strings.TrimSpace(candidate.ID)
		candidate.Name = strings.TrimSpace(candidate.Name)
		if candidate.ID == "" {
			candidate.ID = candidate.Name
		}
		if candidate.Name == "" {
			candidate.Name = candidate.ID
		}
		if candidate.ID == "" || seen[candidate.ID] {
			return
		}
		seen[candidate.ID] = true
		candidates = append(candidates, candidate)
	}

	var parseList func(any)
	parseList = func(value any) {
		switch items := value.(type) {
		case []any:
			for _, item := range items {
				switch typed := item.(type) {
				case string:
					appendCandidate(LLMModelCandidate{ID: typed, Name: typed})
				case map[string]any:
					id := firstString(typed, "id", "name", "model")
					name := firstString(typed, "name", "id", "model")
					appendCandidate(LLMModelCandidate{
						ID:      id,
						Name:    name,
						OwnedBy: firstString(typed, "owned_by", "owner"),
					})
				}
			}
		}
	}

	switch typed := payload.(type) {
	case []any:
		parseList(typed)
	case map[string]any:
		for _, key := range []string{"data", "models", "model_list"} {
			if value, ok := typed[key]; ok {
				parseList(value)
			}
		}
	}

	return candidates
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return typed
			}
		case json.Number:
			return typed.String()
		}
	}
	return ""
}

func SafeLLMProviders(providers []models.LLMProvider) []models.LLMProvider {
	items := make([]models.LLMProvider, 0, len(providers))
	for _, provider := range providers {
		items = append(items, SafeLLMProvider(provider))
	}
	return items
}

func SafeLLMProvider(provider models.LLMProvider) models.LLMProvider {
	apiKey := strings.TrimSpace(provider.APIKey)
	provider.HasAPIKey = apiKey != ""
	provider.APIKeyMasked = maskProviderSecret(apiKey)
	provider.APIKey = ""
	return provider
}

func maskProviderSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
