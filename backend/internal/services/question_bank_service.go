package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

type QuestionBankItem struct {
	ID            uint                   `json:"id"`
	Subject       string                 `json:"subject"`
	Level1        string                 `json:"level1"`
	Level2        string                 `json:"level2"`
	Title         string                 `json:"title"`
	Stem          string                 `json:"stem"`
	Answer        string                 `json:"answer"`
	Difficulty    string                 `json:"difficulty"`
	Tags          []string               `json:"tags"`
	Source        string                 `json:"source"`
	DocumentID    uint                   `json:"document_id,omitempty"`
	Document      string                 `json:"document,omitempty"`
	QuestionNo    string                 `json:"question_no,omitempty"`
	Materials     []QuestionBankMaterial `json:"materials,omitempty"`
	MaterialCount int                    `json:"material_count"`
}

type QuestionBankMaterial struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type QuestionBankService struct {
	db *gorm.DB
}

func NewQuestionBankService(db *gorm.DB) *QuestionBankService {
	return &QuestionBankService{db: db}
}

func (s *QuestionBankService) List(userID uint) ([]QuestionBankItem, error) {
	s.ensureEssayQuestions(userID)

	var essayQuestions []models.EssayQuestion
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&essayQuestions).Error; err != nil {
		return nil, err
	}

	documentTitles := s.essayDocumentTitles(userID, essayQuestions)
	materials := s.materialsForQuestions(userID, essayQuestions)

	items := make([]QuestionBankItem, 0, len(essayQuestions))
	for _, q := range essayQuestions {
		questionMaterials := materials[q.ID]
		items = append(items, QuestionBankItem{
			ID:            q.ID,
			Subject:       "申论",
			Level1:        q.QuestionType,
			Level2:        "材料题",
			Title:         q.Title,
			Stem:          q.QuestionText,
			Answer:        "",
			Difficulty:    "待评估",
			Tags:          []string{"申论", q.QuestionType},
			Source:        "申论解析",
			DocumentID:    q.DocumentID,
			Document:      documentTitles[q.DocumentID],
			QuestionNo:    q.QuestionNo,
			Materials:     questionMaterials,
			MaterialCount: len(questionMaterials),
		})
	}

	var mistakes []models.Mistake
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&mistakes).Error; err != nil {
		return nil, err
	}
	for _, m := range mistakes {
		items = append(items, QuestionBankItem{
			ID:         m.ID,
			Subject:    m.Subject,
			Level1:     m.QuestionType,
			Level2:     m.SubType,
			Title:      m.Title,
			Stem:       m.Stem,
			Answer:     m.CorrectAnswer,
			Difficulty: "待评估",
			Tags:       decodeStringSlice(m.Tags),
			Source:     m.Source,
		})
	}
	return items, nil
}

func (s *QuestionBankService) Get(userID uint, id uint) (*QuestionBankItem, error) {
	s.ensureEssayQuestions(userID)

	var q models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&q).Error; err != nil {
		return nil, err
	}

	documentTitles := s.essayDocumentTitles(userID, []models.EssayQuestion{q})
	materials := s.materialsForQuestions(userID, []models.EssayQuestion{q})
	questionMaterials := materials[q.ID]
	return &QuestionBankItem{
		ID:            q.ID,
		Subject:       "申论",
		Level1:        q.QuestionType,
		Level2:        "材料题",
		Title:         q.Title,
		Stem:          q.QuestionText,
		Answer:        "",
		Difficulty:    "待评估",
		Tags:          []string{"申论", q.QuestionType},
		Source:        "申论解析",
		DocumentID:    q.DocumentID,
		Document:      documentTitles[q.DocumentID],
		QuestionNo:    q.QuestionNo,
		Materials:     questionMaterials,
		MaterialCount: len(questionMaterials),
	}, nil
}

func (s *QuestionBankService) Delete(userID uint, id uint) error {
	var q models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&q).Error; err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND question_id = ?", userID, id).Delete(&models.EssayReview{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND question_id = ?", userID, id).Delete(&models.EssaySectionRelation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND question_id = ?", userID, id).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND id = ?", userID, id).Delete(&models.EssayQuestion{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("question not found")
		}
		return tx.Model(&models.EssayDocument{}).
			Where("user_id = ? AND id = ?", userID, q.DocumentID).
			Update("status", "parsed").Error
	})
}

func (s *QuestionBankService) Update(userID uint, id uint, input QuestionBankItem) (*QuestionBankItem, error) {
	var q models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&q).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Title) != "" {
		q.Title = strings.TrimSpace(input.Title)
	}
	if strings.TrimSpace(input.Level1) != "" {
		q.QuestionType = strings.TrimSpace(input.Level1)
	}
	if strings.TrimSpace(input.Stem) != "" {
		q.QuestionText = strings.TrimSpace(input.Stem)
	}
	if strings.TrimSpace(input.QuestionNo) != "" {
		q.QuestionNo = strings.TrimSpace(input.QuestionNo)
	}
	if err := s.db.Save(&q).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, id)
}

func (s *QuestionBankService) ensureEssayQuestions(userID uint) {
	var documents []models.EssayDocument
	if err := s.db.Where("user_id = ?", userID).Find(&documents).Error; err != nil {
		return
	}
	essayService := NewEssayService(s.db)
	for _, document := range documents {
		var sectionCount int64
		if err := s.db.Model(&models.EssaySection{}).
			Where("user_id = ? AND document_id = ?", userID, document.ID).
			Count(&sectionCount).Error; err != nil || sectionCount == 0 {
			continue
		}
		var questionCount int64
		if err := s.db.Model(&models.EssayQuestion{}).
			Where("user_id = ? AND document_id = ?", userID, document.ID).
			Count(&questionCount).Error; err != nil || questionCount > 0 {
			continue
		}
		_, _ = essayService.AssembleQuestions(userID, document.ID)
	}
}

func (s *QuestionBankService) essayDocumentTitles(userID uint, questions []models.EssayQuestion) map[uint]string {
	documentIDs := make([]uint, 0)
	seen := make(map[uint]bool)
	for _, q := range questions {
		if q.DocumentID == 0 || seen[q.DocumentID] {
			continue
		}
		seen[q.DocumentID] = true
		documentIDs = append(documentIDs, q.DocumentID)
	}

	titles := make(map[uint]string)
	if len(documentIDs) == 0 {
		return titles
	}
	var documents []models.EssayDocument
	if err := s.db.Where("user_id = ? AND id IN ?", userID, documentIDs).Find(&documents).Error; err != nil {
		return titles
	}
	for _, document := range documents {
		titles[document.ID] = document.Title
	}
	return titles
}

func (s *QuestionBankService) materialsForQuestions(userID uint, questions []models.EssayQuestion) map[uint][]QuestionBankMaterial {
	result := make(map[uint][]QuestionBankMaterial)
	if len(questions) == 0 {
		return result
	}

	questionIDs := make([]uint, 0, len(questions))
	for _, q := range questions {
		questionIDs = append(questionIDs, q.ID)
	}

	var relations []models.EssaySectionRelation
	_ = s.db.Where("user_id = ? AND question_id IN ? AND relation_type = ?", userID, questionIDs, "question_material").Find(&relations).Error
	sectionIDs := make([]uint, 0, len(relations))
	for _, rel := range relations {
		sectionIDs = append(sectionIDs, rel.SectionID)
	}

	sectionByID := make(map[uint]models.EssaySection)
	if len(sectionIDs) > 0 {
		var sections []models.EssaySection
		if err := s.db.Where("user_id = ? AND id IN ?", userID, sectionIDs).Find(&sections).Error; err == nil {
			for _, section := range sections {
				sectionByID[section.ID] = section
			}
		}
	}

	for _, rel := range relations {
		section, ok := sectionByID[rel.SectionID]
		if !ok {
			continue
		}
		result[rel.QuestionID] = append(result[rel.QuestionID], questionBankMaterial(section))
	}

	for _, q := range questions {
		if len(result[q.ID]) > 0 || q.DocumentID == 0 {
			continue
		}
		var sections []models.EssaySection
		if err := s.db.Where("user_id = ? AND document_id = ? AND section_type = ?", userID, q.DocumentID, "material").Order("id asc").Find(&sections).Error; err != nil {
			continue
		}
		for _, section := range sections {
			result[q.ID] = append(result[q.ID], questionBankMaterial(section))
		}
		if len(result[q.ID]) == 0 {
			result[q.ID] = []QuestionBankMaterial{{
				Title:   "未关联材料",
				Content: fmt.Sprintf("题目 %s 尚未匹配到材料，请重新解析文档或检查材料区切分。", strings.TrimSpace(q.QuestionNo)),
			}}
		}
	}

	return result
}

func questionBankMaterial(section models.EssaySection) QuestionBankMaterial {
	title := strings.TrimSpace(section.Title)
	if title == "" {
		title = fmt.Sprintf("材料 #%d", section.ID)
	}
	return QuestionBankMaterial{
		ID:      section.ID,
		Title:   title,
		Content: section.Content,
	}
}

func decodeStringSlice(value string) []string {
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err == nil {
		return items
	}
	if value == "" {
		return nil
	}
	return []string{value}
}
