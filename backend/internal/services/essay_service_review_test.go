package services

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gkweb/backend/internal/models"
)

func TestReviewAnswerStoresContextAndPromptSnapshot(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.PromptTemplate{},
		&models.EssayDocument{},
		&models.EssayQuestion{},
		&models.EssaySection{},
		&models.EssaySectionRelation{},
		&models.EssayReview{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	userID := uint(1)
	prompt := models.PromptTemplate{
		BaseModel:    models.BaseModel{UserID: userID},
		TaskType:     "申论批改",
		Name:         "测试批改模板",
		SystemPrompt: "系统模板 {{question_type}}",
		UserPrompt:   "题目：{{question_text}}\n材料：{{materials}}\n参考：{{reference_answers}}\n答案：{{answer}}",
		Version:      "v-test",
		Enabled:      true,
	}
	if err := db.Create(&prompt).Error; err != nil {
		t.Fatalf("create prompt: %v", err)
	}
	document := models.EssayDocument{
		BaseModel:    models.BaseModel{UserID: userID},
		Title:        "申论套题",
		DocumentRole: "combined",
	}
	if err := db.Create(&document).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	question := models.EssayQuestion{
		BaseModel:      models.BaseModel{UserID: userID},
		DocumentID:     document.ID,
		Title:          "第一题",
		QuestionNo:     "1",
		QuestionType:   "归纳概括题",
		QuestionText:   "概括材料中的问题。",
		MaxScore:       20,
		WordLimit:      300,
		CustomPromptID: &prompt.ID,
	}
	if err := db.Create(&question).Error; err != nil {
		t.Fatalf("create question: %v", err)
	}
	material := models.EssaySection{
		BaseModel:   models.BaseModel{UserID: userID},
		DocumentID:  document.ID,
		SectionType: "material",
		Content:     "材料内容 A",
	}
	answer := models.EssaySection{
		BaseModel:   models.BaseModel{UserID: userID},
		DocumentID:  document.ID,
		SectionType: "answer",
		Content:     "参考答案 B",
	}
	if err := db.Create(&material).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	if err := db.Create(&answer).Error; err != nil {
		t.Fatalf("create answer: %v", err)
	}
	relations := []models.EssaySectionRelation{
		{BaseModel: models.BaseModel{UserID: userID}, DocumentID: document.ID, QuestionID: question.ID, SectionID: material.ID, RelationType: "question_material"},
		{BaseModel: models.BaseModel{UserID: userID}, DocumentID: document.ID, QuestionID: question.ID, SectionID: answer.ID, RelationType: "question_answer"},
	}
	if err := db.Create(&relations).Error; err != nil {
		t.Fatalf("create relations: %v", err)
	}

	result, err := NewEssayService(db).ReviewAnswer(userID, question.ID, 0, "我的答案")
	if err != nil {
		t.Fatalf("review answer: %v", err)
	}
	review := result.Review
	if review.PromptTemplateID == nil || *review.PromptTemplateID != prompt.ID {
		t.Fatalf("prompt template snapshot missing: %#v", review)
	}
	for label, text := range map[string]string{
		"question snapshot": review.QuestionSnapshot,
		"material snapshot": review.MaterialSnapshot,
		"answer snapshot":   review.AnswerSnapshot,
		"system prompt":     review.SystemPromptSnapshot,
		"user prompt":       review.UserPromptSnapshot,
	} {
		if strings.TrimSpace(text) == "" {
			t.Fatalf("%s was empty", label)
		}
	}
	if !strings.Contains(review.UserPromptSnapshot, "材料内容 A") || !strings.Contains(review.UserPromptSnapshot, "参考答案 B") || !strings.Contains(review.UserPromptSnapshot, "我的答案") {
		t.Fatalf("user prompt variables were not replaced: %s", review.UserPromptSnapshot)
	}
	if !strings.Contains(review.QuestionSnapshot, "概括材料中的问题") {
		t.Fatalf("question snapshot missing question text: %s", review.QuestionSnapshot)
	}
}
