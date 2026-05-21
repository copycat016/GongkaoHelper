package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
	"gkweb/backend/internal/parser"
)

const (
	EssayChunkMaterial        = "material"
	EssayChunkQuestion        = "question"
	EssayChunkReferenceAnswer = "reference_answer"
	EssayChunkScoringRule     = "scoring_rule"
	EssayChunkExplanation     = "explanation"
	EssayChunkUnknown         = "unknown"
)

type EssayService struct {
	db *gorm.DB
}

type EssayReviewResult struct {
	Review      models.EssayReview `json:"review"`
	Score       float64            `json:"score"`
	MaxScore    int                `json:"max_score"`
	Summary     string             `json:"summary"`
	Suggestions []string           `json:"suggestions"`
	Context     ReviewContext      `json:"context"`
}

type BoundaryDebugResult struct {
	DocumentID    uint                 `json:"document_id"`
	ModelID       uint                 `json:"model_id"`
	BlockCount    int                  `json:"block_count"` // 新流程中表示行数
	Prompt        string               `json:"prompt"`
	RawResponse   string               `json:"raw_response"`
	ExtractedJSON string               `json:"extracted_json"`
	Plan          *parser.BoundaryPlan `json:"plan,omitempty"`
	ParseError    string               `json:"parse_error,omitempty"`
	ApplyError    string               `json:"apply_error,omitempty"`
	Sections      []parser.Section     `json:"sections,omitempty"`
}

// ReviewContext 展示题目批改时实际使用的上下文（非整篇 PDF）。
type ReviewContext struct {
	QuestionText string   `json:"question_text"`
	Materials    []string `json:"materials"`
	Answers      []string `json:"answers"`
}

type EssayQuestionPayload struct {
	DocumentID     uint   `json:"document_id"`
	QuestionNo     string `json:"question_no"`
	Title          string `json:"title"`
	QuestionType   string `json:"question_type"`
	QuestionText   string `json:"question_text"`
	MaxScore       int    `json:"max_score"`
	WordLimit      int    `json:"word_limit"`
	ScoringRubric  string `json:"scoring_rubric"`
	CustomPromptID *uint  `json:"custom_prompt_id"`
	Status         string `json:"status"`
}

type EssaySectionPayload struct {
	SectionType        string `json:"section_type"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	QuestionNo         string `json:"question_no"`
	RelatedQuestionNos string `json:"related_question_nos"`
}

func NewEssayService(db *gorm.DB) *EssayService {
	return &EssayService{db: db}
}

func (s *EssayService) ListDocuments(userID uint) ([]models.EssayDocument, error) {
	var documents []models.EssayDocument
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&documents).Error
	return documents, err
}

func (s *EssayService) CreateDocument(document *models.EssayDocument) error {
	if strings.TrimSpace(document.DocumentRole) == "" {
		document.DocumentRole = "combined"
	}
	return s.db.Create(document).Error
}

func (s *EssayService) MarkDocumentFailed(userID uint, documentID uint, message string) {
	var document models.EssayDocument
	if err := s.db.Where("user_id = ? AND id = ?", userID, documentID).First(&document).Error; err != nil {
		return
	}
	document.Status = "parse_failed"
	document.Note = truncateNote(message)
	_ = s.db.Save(&document).Error
}

// ParseDocument 使用新的 parser 模块解析文档，按语义区域（section）存储。
func (s *EssayService) ParseDocument(userID uint, documentID uint, rawText string) (*models.EssayDocument, []models.EssaySection, error) {
	return s.ParseDocumentWithBoundaryModel(userID, documentID, rawText, 0)
}

// ParseDocumentWithBoundaryModel 是核心解析流程。
//
// 新流程（boundaryModelID > 0）：
//
//	PDF文本 → 基础清洗(去空行/页码) → 编行号 → 整份文本交给 LLM 按行号切分 → 存 sections → 自动组装 questions
//
// 旧流程兜底（boundaryModelID == 0）：
//
//	PDF文本 → adapter → blocks → cleaner → 锚点+状态机切分 → 存 sections
func (s *EssayService) ParseDocumentWithBoundaryModel(userID uint, documentID uint, rawText string, boundaryModelID uint) (*models.EssayDocument, []models.EssaySection, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, nil, err
	}
	document.Status = "parsing"
	document.Note = ""
	_ = s.db.Save(document).Error

	parseSource, err := s.documentParseSource(document, rawText)
	if err != nil {
		return nil, nil, err
	}
	rawText = parseSource.Text

	var sections []models.EssaySection

	if boundaryModelID > 0 {
		// ── 新流程：LLM 直接按行号切分 ──
		// 1. 基础清洗 + 编行号
		lines := parser.PrepareNumberedLines(rawText)
		if len(lines) == 0 {
			return nil, nil, errors.New("文档清洗后没有有效文本行")
		}

		// 2. LLM 切分
		plan, err := s.suggestBoundaryPlanFromLines(userID, boundaryModelID, lines)
		if err != nil {
			return nil, nil, fmt.Errorf("LLM 切分失败: %w", err)
		}

		// 3. 按行号范围提取 section
		parsedSections, err := parser.ApplyBoundaryPlanToLines(lines, plan)
		if err != nil {
			return nil, nil, fmt.Errorf("应用切分结果失败: %w", err)
		}

		sections = s.sectionsFromParsedSections(userID, documentID, parsedSections)
		document.ClassModelID = boundaryModelID
	} else {
		// ── 旧流程兜底：规则引擎 ──
		p := parser.NewDefault()
		cleaned, err := p.PrepareString(fmt.Sprintf("doc_%d", documentID), rawText)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare document failed: %w", err)
		}
		result, err := p.ParsePrepared(cleaned)
		if err != nil {
			return nil, nil, fmt.Errorf("parser failed: %w", err)
		}
		sections = s.sectionsFromResult(userID, documentID, result)
	}

	sections = normalizeEssayQuestionSections(sections)
	if len(sections) == 0 {
		return nil, nil, errors.New("no sections parsed from document")
	}

	// 4. 保存 sections 并自动组装 questions
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 清理旧数据
		for _, table := range []any{&models.EssaySectionRelation{}, &models.EssayQuestionChunk{}, &models.EssayQuestion{}, &models.EssaySection{}, &models.EssayChunk{}} {
			if err := tx.Where("user_id = ? AND document_id = ?", userID, document.ID).Delete(table).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&sections).Error; err != nil {
			return err
		}
		document.Status = "parsed"
		document.ChunkCount = len(sections)
		document.PageCount = maxSectionPageEnd(sections)
		document.Note = ""
		return tx.Save(document).Error
	})
	if err != nil {
		return nil, nil, err
	}

	// 5. 自动组装 questions（解析成功后直接进入申论题库）
	if _, assembleErr := s.AssembleQuestions(userID, documentID); assembleErr != nil {
		// 组装失败不影响解析结果，只记录到 note
		document.Note = "自动组装题目失败: " + assembleErr.Error()
		_ = s.db.Save(document).Error
	}

	return document, sections, nil
}

// suggestBoundaryPlanFromLines 用新流程调用 LLM：整份文本 + 行号 → JSON 边界。
func (s *EssayService) suggestBoundaryPlanFromLines(userID uint, modelID uint, lines []parser.NumberedLine) (parser.BoundaryPlan, error) {
	model, provider, err := s.getModelProvider(userID, modelID)
	if err != nil {
		return parser.BoundaryPlan{}, err
	}
	if !model.Enabled || !provider.Enabled {
		return parser.BoundaryPlan{}, errors.New("boundary model or provider is disabled")
	}

	prompt := parser.BuildBoundaryPromptFromLines(lines)
	systemPrompt := parser.BuildBoundarySystemPrompt()
	content, err := callOpenAICompatibleChat(provider.BaseURL, provider.APIKey, model.Name, systemPrompt, prompt)
	if err != nil {
		return parser.BoundaryPlan{}, err
	}
	content = extractJSONContent(content)

	var plan parser.BoundaryPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return parser.BoundaryPlan{}, fmt.Errorf("decode boundary json failed: %w\nraw: %s", err, truncateForError(content, 500))
	}
	if len(plan.Sections) == 0 {
		return parser.BoundaryPlan{}, errors.New("boundary json has no sections")
	}
	return plan, nil
}

// sectionsFromParsedSections 将 parser.Section 转为 models.EssaySection。
func (s *EssayService) sectionsFromParsedSections(userID uint, documentID uint, parsedSections []parser.Section) []models.EssaySection {
	sections := make([]models.EssaySection, 0, len(parsedSections))
	for _, sec := range parsedSections {
		relatedQNos := ""
		if len(sec.RelatedQuestionNos) > 0 {
			relatedQNos = strings.Join(sec.RelatedQuestionNos, ",")
		}
		sections = append(sections, models.EssaySection{
			BaseModel:          models.BaseModel{UserID: userID},
			DocumentID:         documentID,
			PageStart:          sec.PageStart,
			PageEnd:            sec.PageEnd,
			SectionType:        string(sec.Type),
			Title:              sec.Title,
			Content:            sec.Text,
			QuestionNo:         sec.QuestionNo,
			RelatedQuestionNos: relatedQNos,
			Confidence:         sec.Confidence,
			Reason:             sec.Reason,
			ParsedByLLM:        true,
		})
	}
	return sections
}

func (s *EssayService) DebugBoundarySplit(userID uint, documentID uint, rawText string, modelID uint) (*BoundaryDebugResult, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, err
	}
	parseSource, err := s.documentParseSource(document, rawText)
	if err != nil {
		return nil, err
	}
	rawText = parseSource.Text

	// 新流程：按行号切分
	lines := parser.PrepareNumberedLines(rawText)
	prompt := parser.BuildBoundaryPromptFromLines(lines)

	result := &BoundaryDebugResult{
		DocumentID: documentID,
		ModelID:    modelID,
		BlockCount: len(lines), // 复用字段，表示行数
		Prompt:     prompt,
	}
	if modelID == 0 {
		result.ParseError = "boundary_model_id is required for LLM debug"
		return result, nil
	}

	model, provider, err := s.getModelProvider(userID, modelID)
	if err != nil {
		return nil, err
	}
	systemPrompt := parser.BuildBoundarySystemPrompt()
	rawResponse, content, err := callOpenAICompatibleChatDebug(provider.BaseURL, provider.APIKey, model.Name, systemPrompt, prompt)
	if err != nil {
		result.ParseError = err.Error()
		return result, nil
	}
	result.RawResponse = rawResponse
	result.ExtractedJSON = extractJSONContent(content)

	var plan parser.BoundaryPlan
	if err := json.Unmarshal([]byte(result.ExtractedJSON), &plan); err != nil {
		result.ParseError = err.Error()
		return result, nil
	}
	result.Plan = &plan

	sections, err := parser.ApplyBoundaryPlanToLines(lines, plan)
	if err != nil {
		result.ApplyError = err.Error()
		return result, nil
	}
	result.Sections = sections
	return result, nil
}

func (s *EssayService) documentParseSource(document *models.EssayDocument, rawText string) (*DocumentParseResult, error) {
	if strings.TrimSpace(rawText) != "" || strings.TrimSpace(document.FilePath) != "" {
		return ParseDocumentSource(DocumentParseInput{
			RawText: rawText,
			PDFPath: document.FilePath,
		})
	}
	return parseRawTextDocument(sampleEssayPDFText(document.Title)), nil
}

func (s *EssayService) getModelProvider(userID uint, modelID uint) (models.LLMModel, models.LLMProvider, error) {
	var model models.LLMModel
	if err := s.db.Where("user_id = ? AND id = ?", userID, modelID).First(&model).Error; err != nil {
		return models.LLMModel{}, models.LLMProvider{}, err
	}
	var provider models.LLMProvider
	if err := s.db.Where("user_id = ? AND id = ?", userID, model.ProviderID).First(&provider).Error; err != nil {
		return models.LLMModel{}, models.LLMProvider{}, err
	}
	return model, provider, nil
}

func callOpenAICompatibleChat(baseURL string, apiKey string, model string, systemPrompt string, userPrompt string) (string, error) {
	_, content, err := callOpenAICompatibleChatDebug(baseURL, apiKey, model, systemPrompt, userPrompt)
	return content, err
}

func callOpenAICompatibleChatDebug(baseURL string, apiKey string, model string, systemPrompt string, userPrompt string) (string, string, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return "", "", errors.New("llm base_url is required")
	}
	if !strings.HasSuffix(base, "/chat/completions") {
		base += "/chat/completions"
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1,
		"max_tokens":  4096,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, base, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	rawBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", "", readErr
	}
	rawResponse := string(rawBody)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return rawResponse, "", fmt.Errorf("llm api returned %d: %s", resp.StatusCode, truncateForError(rawResponse, 1024))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return rawResponse, "", err
	}
	if len(result.Choices) == 0 {
		return rawResponse, "", errors.New("llm returned no choices")
	}
	return rawResponse, result.Choices[0].Message.Content, nil
}

func extractJSONContent(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		var kept []string
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				continue
			}
			kept = append(kept, line)
		}
		return strings.TrimSpace(strings.Join(kept, "\n"))
	}
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end >= start {
		return strings.TrimSpace(content[start : end+1])
	}
	return content
}

func truncateForError(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}

func (s *EssayService) sectionsFromResult(userID uint, documentID uint, result parser.StructuredResult) []models.EssaySection {
	sections := make([]models.EssaySection, 0, len(result.Sections))
	for _, sec := range result.Sections {
		blockIDs := make([]string, len(sec.Blocks))
		for i, b := range sec.Blocks {
			blockIDs[i] = b.ID
		}
		relatedQNos := ""
		if len(sec.RelatedQuestionNos) > 0 {
			relatedQNos = strings.Join(sec.RelatedQuestionNos, ",")
		}
		sections = append(sections, models.EssaySection{
			BaseModel:          models.BaseModel{UserID: userID},
			DocumentID:         documentID,
			PageStart:          sec.PageStart,
			PageEnd:            sec.PageEnd,
			SectionType:        string(sec.Type),
			Title:              sec.Title,
			Content:            sec.Text,
			BlockIDs:           strings.Join(blockIDs, ","),
			QuestionNo:         sec.QuestionNo,
			RelatedQuestionNos: relatedQNos,
			Confidence:         sec.Confidence,
			Reason:             sec.Reason,
		})
	}
	return sections
}

// ListSections 返回文档的语义区域（替代旧版 ListChunks）。
func (s *EssayService) ListSections(userID uint, documentID uint) ([]models.EssaySection, error) {
	var sections []models.EssaySection
	err := s.db.Where("user_id = ? AND document_id = ?", userID, documentID).Order("id asc").Find(&sections).Error
	return sections, err
}

// ListChunks 保留向后兼容：从 sections 转换输出（旧表不再写入）。
func (s *EssayService) ListChunks(userID uint, documentID uint) ([]models.EssayChunk, error) {
	sections, err := s.ListSections(userID, documentID)
	if err != nil {
		return nil, err
	}
	chunks := make([]models.EssayChunk, 0, len(sections))
	for i, sec := range sections {
		chunks = append(chunks, models.EssayChunk{
			BaseModel:          models.BaseModel{UserID: sec.UserID},
			DocumentID:         sec.DocumentID,
			PageNo:             sec.PageStart,
			ChunkIndex:         i + 1,
			Content:            sec.Content,
			ChunkType:          sec.SectionType,
			Confidence:         sec.Confidence,
			ClassificationNote: sec.Reason,
		})
	}
	return chunks, nil
}

// ClassifyChunks 在新流程中已无需单独调用（parser 已在解析时完成分类），
// 保留方法以兼容旧接口，直接返回当前 sections 的 chunk 视图。
func (s *EssayService) ClassifyChunks(userID uint, documentID uint, modelID uint) ([]models.EssayChunk, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, err
	}
	chunks, err := s.ListChunks(userID, documentID)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, errors.New("document has no sections, parse it first")
	}

	document.Status = "classified"
	document.ClassModelID = modelID
	_ = s.db.Save(document).Error
	return chunks, nil
}

// AssembleQuestions 基于语义区域（section）组装题目和关联关系。
// 关联策略：
// 1. 优先使用 section 中的 question_no / related_question_nos 精确关联
// 2. 答案数量与题目数量相同时按顺序一一对应
// 3. 兜底：所有材料关联到所有题目
func (s *EssayService) AssembleQuestions(userID uint, documentID uint) ([]models.EssayQuestion, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, err
	}

	sections, err := s.ListSections(userID, documentID)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return nil, errors.New("document has no sections")
	}

	var questions []models.EssayQuestion
	var relations []models.EssaySectionRelation

	// 按类型分组
	type sectionRef struct {
		ID                 uint
		QuestionNo         string
		RelatedQuestionNos []string
	}
	var materialRefs []sectionRef
	var answerRefs []sectionRef

	for _, sec := range sections {
		relQNos := parseCommaSeparated(sec.RelatedQuestionNos)
		switch sec.SectionType {
		case string(parser.SectionMaterial):
			materialRefs = append(materialRefs, sectionRef{ID: sec.ID, RelatedQuestionNos: relQNos})
		case string(parser.SectionAnswer):
			answerRefs = append(answerRefs, sectionRef{ID: sec.ID, QuestionNo: sec.QuestionNo, RelatedQuestionNos: relQNos})
		case string(parser.SectionQuestion):
			qNo := sec.QuestionNo
			if qNo == "" {
				qNo = fmt.Sprintf("%d", len(questions)+1)
			}
			q := models.EssayQuestion{
				BaseModel:    models.BaseModel{UserID: userID},
				DocumentID:   documentID,
				Title:        questionTitle(document.Title, sec.Content, qNo, len(questions)+1),
				QuestionNo:   qNo,
				QuestionType: guessQuestionType(sec.Content),
				QuestionText: sec.Content,
				MaxScore:     guessMaxScore(sec.Content),
				WordLimit:    guessWordLimit(sec.Content),
				Status:       "assembled",
			}
			questions = append(questions, q)
		}
	}

	if len(questions) == 0 {
		return nil, errors.New("no question sections found, parse document first")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssaySectionRelation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestion{}).Error; err != nil {
			return err
		}

		// 先创建所有 question 以获得 ID
		for index := range questions {
			if err := tx.Create(&questions[index]).Error; err != nil {
				return err
			}
		}

		// question_no -> question DB ID
		qNoToID := make(map[string]uint)
		for _, q := range questions {
			qNoToID[q.QuestionNo] = q.ID
		}

		// ── 关联材料 ──
		// 材料通常关联到所有题目（申论材料是共享的），除非 related_question_nos 明确指定
		for _, mat := range materialRefs {
			if len(mat.RelatedQuestionNos) > 0 {
				// 精确关联
				for _, qNo := range mat.RelatedQuestionNos {
					if qID, ok := qNoToID[qNo]; ok {
						relations = append(relations, models.EssaySectionRelation{
							BaseModel: models.BaseModel{UserID: userID}, DocumentID: documentID,
							QuestionID: qID, SectionID: mat.ID, RelationType: "question_material",
						})
					}
				}
			} else {
				// 保守策略：关联到所有题目
				for _, q := range questions {
					relations = append(relations, models.EssaySectionRelation{
						BaseModel: models.BaseModel{UserID: userID}, DocumentID: documentID,
						QuestionID: q.ID, SectionID: mat.ID, RelationType: "question_material",
					})
				}
			}
		}

		// ── 关联答案 ──
		linkedAnswers := make(map[int]bool)
		// 优先使用 related_question_nos 精确关联
		for aIdx, ans := range answerRefs {
			relNos := ans.RelatedQuestionNos
			// 如果 RelatedQuestionNos 为空但 QuestionNo 不为空，也作为关联依据
			if len(relNos) == 0 && ans.QuestionNo != "" {
				relNos = []string{ans.QuestionNo}
			}
			if len(relNos) > 0 {
				for _, qNo := range relNos {
					if qID, ok := qNoToID[qNo]; ok {
						relations = append(relations, models.EssaySectionRelation{
							BaseModel: models.BaseModel{UserID: userID}, DocumentID: documentID,
							QuestionID: qID, SectionID: ans.ID, RelationType: "question_answer",
						})
						linkedAnswers[aIdx] = true
					}
				}
			}
		}

		// 未关联的答案按顺序对应或保守关联
		var unlinked []int
		for aIdx := range answerRefs {
			if !linkedAnswers[aIdx] {
				unlinked = append(unlinked, aIdx)
			}
		}
		if len(unlinked) > 0 {
			if len(unlinked) == len(questions) {
				// 按顺序一一对应
				for i, aIdx := range unlinked {
					relations = append(relations, models.EssaySectionRelation{
						BaseModel: models.BaseModel{UserID: userID}, DocumentID: documentID,
						QuestionID: questions[i].ID, SectionID: answerRefs[aIdx].ID, RelationType: "question_answer",
					})
				}
			} else {
				// 兜底：关联到所有题目
				for _, aIdx := range unlinked {
					for _, q := range questions {
						relations = append(relations, models.EssaySectionRelation{
							BaseModel: models.BaseModel{UserID: userID}, DocumentID: documentID,
							QuestionID: q.ID, SectionID: answerRefs[aIdx].ID, RelationType: "question_answer",
						})
					}
				}
			}
		}

		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		document.Status = "assembled"
		return tx.Save(document).Error
	})
	if err != nil {
		return nil, err
	}

	return questions, nil
}

// parseCommaSeparated 将逗号分隔的字符串拆分为非空字符串切片。
func parseCommaSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func normalizeQuestionDefaults(question *models.EssayQuestion) {
	if strings.TrimSpace(question.QuestionText) == "" {
		question.QuestionText = "待补充题面"
	}
	if strings.TrimSpace(question.Title) == "" {
		question.Title = questionTitle("申论题", question.QuestionText, question.QuestionNo, 1)
	}
	if question.MaxScore <= 0 {
		question.MaxScore = 100
	}
	if question.WordLimit <= 0 {
		question.WordLimit = 500
	}
	if strings.TrimSpace(question.Status) == "" {
		question.Status = "assembled"
	}
}

func uniqueUintIDs(ids []uint) []uint {
	seen := make(map[uint]bool)
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

func (s *EssayService) validateSectionForRelation(userID uint, documentID uint, sectionID uint, allowedTypes []string) error {
	var section models.EssaySection
	if err := s.db.Where("user_id = ? AND document_id = ? AND id = ?", userID, documentID, sectionID).First(&section).Error; err != nil {
		return err
	}
	for _, allowed := range allowedTypes {
		if section.SectionType == allowed {
			return nil
		}
	}
	return fmt.Errorf("section %d type %s cannot be used for this relation", sectionID, section.SectionType)
}

func (s *EssayService) attachQuestionRelationIDs(userID uint, questions []models.EssayQuestion) []models.EssayQuestion {
	if len(questions) == 0 {
		return questions
	}
	ids := make([]uint, 0, len(questions))
	indexByID := make(map[uint]int, len(questions))
	for index, question := range questions {
		ids = append(ids, question.ID)
		indexByID[question.ID] = index
	}
	var relations []models.EssaySectionRelation
	if err := s.db.Where("user_id = ? AND question_id IN ?", userID, ids).Find(&relations).Error; err != nil {
		return questions
	}
	for _, rel := range relations {
		index, ok := indexByID[rel.QuestionID]
		if !ok {
			continue
		}
		switch rel.RelationType {
		case "question_material":
			questions[index].MaterialSectionIDs = append(questions[index].MaterialSectionIDs, rel.SectionID)
		case "question_answer":
			questions[index].AnswerSectionIDs = append(questions[index].AnswerSectionIDs, rel.SectionID)
		}
	}
	return questions
}

func (s *EssayService) ListQuestions(userID uint, documentID uint) ([]models.EssayQuestion, error) {
	var questions []models.EssayQuestion
	if err := s.db.Where("user_id = ? AND document_id = ?", userID, documentID).Order("id asc").Find(&questions).Error; err != nil {
		return nil, err
	}
	return s.attachQuestionRelationIDs(userID, questions), nil
}

func (s *EssayService) CreateQuestion(userID uint, payload EssayQuestionPayload) (*models.EssayQuestion, error) {
	if payload.DocumentID == 0 {
		return nil, errors.New("document_id is required")
	}
	if _, err := s.getDocument(userID, payload.DocumentID); err != nil {
		return nil, err
	}
	question := models.EssayQuestion{
		BaseModel:      models.BaseModel{UserID: userID},
		DocumentID:     payload.DocumentID,
		QuestionNo:     strings.TrimSpace(payload.QuestionNo),
		Title:          strings.TrimSpace(payload.Title),
		QuestionType:   strings.TrimSpace(payload.QuestionType),
		QuestionText:   strings.TrimSpace(payload.QuestionText),
		MaxScore:       payload.MaxScore,
		WordLimit:      payload.WordLimit,
		Status:         strings.TrimSpace(payload.Status),
		ManuallyEdited: true,
		CustomPromptID: payload.CustomPromptID,
		ScoringRubric:  strings.TrimSpace(payload.ScoringRubric),
	}
	normalizeQuestionDefaults(&question)
	if err := s.db.Create(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *EssayService) UpdateQuestion(userID uint, questionID uint, payload EssayQuestionPayload) (*models.EssayQuestion, error) {
	var question models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, questionID).First(&question).Error; err != nil {
		return nil, err
	}
	question.QuestionNo = strings.TrimSpace(payload.QuestionNo)
	question.Title = strings.TrimSpace(payload.Title)
	question.QuestionType = strings.TrimSpace(payload.QuestionType)
	question.QuestionText = strings.TrimSpace(payload.QuestionText)
	question.MaxScore = payload.MaxScore
	question.WordLimit = payload.WordLimit
	question.CustomPromptID = payload.CustomPromptID
	question.ScoringRubric = strings.TrimSpace(payload.ScoringRubric)
	if strings.TrimSpace(payload.Status) != "" {
		question.Status = strings.TrimSpace(payload.Status)
	}
	question.ManuallyEdited = true
	normalizeQuestionDefaults(&question)
	if err := s.db.Save(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *EssayService) DeleteQuestion(userID uint, questionID uint) error {
	var question models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, questionID).First(&question).Error; err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.EssayReview{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.EssaySectionRelation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND id = ?", userID, questionID).Delete(&models.EssayQuestion{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		_ = question
		return nil
	})
}

func (s *EssayService) UpdateSection(userID uint, sectionID uint, payload EssaySectionPayload) (*models.EssaySection, error) {
	var section models.EssaySection
	if err := s.db.Where("user_id = ? AND id = ?", userID, sectionID).First(&section).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.SectionType) != "" {
		section.SectionType = strings.TrimSpace(payload.SectionType)
	}
	section.Title = strings.TrimSpace(payload.Title)
	section.Content = strings.TrimSpace(payload.Content)
	section.QuestionNo = strings.TrimSpace(payload.QuestionNo)
	section.RelatedQuestionNos = strings.TrimSpace(payload.RelatedQuestionNos)
	if section.Content == "" {
		return nil, errors.New("section content is required")
	}
	if err := s.db.Save(&section).Error; err != nil {
		return nil, err
	}
	return &section, nil
}

func (s *EssayService) ReplaceQuestionRelations(userID uint, questionID uint, materialIDs []uint, answerIDs []uint) (*models.EssayQuestion, error) {
	var question models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, questionID).First(&question).Error; err != nil {
		return nil, err
	}
	relations := make([]models.EssaySectionRelation, 0, len(materialIDs)+len(answerIDs))
	for _, sectionID := range uniqueUintIDs(materialIDs) {
		if err := s.validateSectionForRelation(userID, question.DocumentID, sectionID, []string{string(parser.SectionMaterial)}); err != nil {
			return nil, err
		}
		relations = append(relations, models.EssaySectionRelation{
			BaseModel:    models.BaseModel{UserID: userID},
			DocumentID:   question.DocumentID,
			QuestionID:   question.ID,
			SectionID:    sectionID,
			RelationType: "question_material",
		})
	}
	for _, sectionID := range uniqueUintIDs(answerIDs) {
		if err := s.validateSectionForRelation(userID, question.DocumentID, sectionID, []string{string(parser.SectionAnswer), string(parser.SectionAnalysis)}); err != nil {
			return nil, err
		}
		relations = append(relations, models.EssaySectionRelation{
			BaseModel:    models.BaseModel{UserID: userID},
			DocumentID:   question.DocumentID,
			QuestionID:   question.ID,
			SectionID:    sectionID,
			RelationType: "question_answer",
		})
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.EssaySectionRelation{}).Error; err != nil {
			return err
		}
		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		question.ManuallyEdited = true
		return tx.Save(&question).Error
	}); err != nil {
		return nil, err
	}
	questions := s.attachQuestionRelationIDs(userID, []models.EssayQuestion{question})
	return &questions[0], nil
}

// ReviewAnswer 根据题号只取对应 question + related materials + answer 进行批改。
// 如果 modelID > 0，调用 LLM 进行真实批改；否则使用简单启发式评分。
func (s *EssayService) ReviewAnswer(userID uint, questionID uint, modelID uint, userAnswer string) (*EssayReviewResult, error) {
	var question models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, questionID).First(&question).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(userAnswer) == "" {
		return nil, errors.New("user_answer is required")
	}

	// 获取关联的 sections（材料 + 答案）
	var relations []models.EssaySectionRelation
	if err := s.db.Where("user_id = ? AND question_id = ?", userID, questionID).Find(&relations).Error; err != nil {
		return nil, err
	}

	var materialTexts []string
	var answerTexts []string
	for _, rel := range relations {
		var sec models.EssaySection
		if err := s.db.Where("user_id = ? AND id = ?", userID, rel.SectionID).First(&sec).Error; err != nil {
			continue
		}
		if rel.RelationType == "question_material" {
			materialTexts = append(materialTexts, sec.Content)
		} else if rel.RelationType == "question_answer" {
			answerTexts = append(answerTexts, sec.Content)
		}
	}

	ctx := ReviewContext{
		QuestionText: question.QuestionText,
		Materials:    materialTexts,
		Answers:      answerTexts,
	}

	var score float64
	var resultPayload map[string]any

	if modelID > 0 {
		// ── 调用 LLM 真实批改 ──
		llmResult, err := s.callLLMReview(userID, modelID, question, materialTexts, answerTexts, userAnswer)
		if err != nil {
			// LLM 调用失败时，降级为启发式评分并记录错误
			score = mockEssayScore(userAnswer, question.WordLimit)
			resultPayload = map[string]any{
				"summary":     fmt.Sprintf("LLM 批改调用失败，已降级为启发式评分。错误: %s", err.Error()),
				"suggestions": []string{"请检查 LLM 模型配置是否正确。"},
				"llm_error":   err.Error(),
			}
		} else {
			score = llmResult.Score
			resultPayload = map[string]any{
				"summary":        llmResult.Summary,
				"suggestions":    llmResult.Suggestions,
				"dimensions":     llmResult.Dimensions,
				"scoring_detail": llmResult.ScoringDetail,
				"highlights":     llmResult.Highlights,
			}
		}
	} else {
		// ── 无模型时使用启发式评分 ──
		score = mockEssayScore(userAnswer, question.WordLimit)
		resultPayload = map[string]any{
			"summary": fmt.Sprintf("启发式评分（未选择批改模型）。已提取 %d 段相关材料、%d 段参考答案/评分标准。", len(materialTexts), len(answerTexts)),
			"suggestions": []string{
				"选择批改模型后可获得 LLM 分维度详细批改。",
			},
		}
	}

	resultBytes, _ := json.Marshal(resultPayload)

	review := models.EssayReview{
		BaseModel:     models.BaseModel{UserID: userID},
		QuestionID:    questionID,
		ReviewModelID: modelID,
		UserAnswer:    userAnswer,
		Score:         score,
		MaxScore:      question.MaxScore,
		ResultJSON:    string(resultBytes),
	}
	if err := s.db.Create(&review).Error; err != nil {
		return nil, err
	}

	summary, _ := resultPayload["summary"].(string)
	suggestions, _ := resultPayload["suggestions"].([]string)

	return &EssayReviewResult{
		Review:      review,
		Score:       score,
		MaxScore:    question.MaxScore,
		Summary:     summary,
		Suggestions: suggestions,
		Context:     ctx,
	}, nil
}

// LLMReviewResult 是 LLM 批改的结构化输出。
type LLMReviewResult struct {
	Score         float64           `json:"score"`
	Summary       string            `json:"summary"`
	Suggestions   []string          `json:"suggestions"`
	Dimensions    []ReviewDimension `json:"dimensions"`
	ScoringDetail string            `json:"scoring_detail"`
	Highlights    []ReviewHighlight `json:"highlights"`
}

// ReviewDimension 是一个评分维度。
type ReviewDimension struct {
	Name     string  `json:"name"` // 如 "内容要点", "逻辑结构", "语言表达"
	Score    float64 `json:"score"`
	MaxScore float64 `json:"max_score"`
	Comment  string  `json:"comment"`
}

// ReviewHighlight 标注答案中的亮点或问题。
type ReviewHighlight struct {
	Type    string `json:"type"`    // "good" / "issue"
	Text    string `json:"text"`    // 原文摘录
	Comment string `json:"comment"` // 批注
}

// callLLMReview 构建批改 prompt 并调用 LLM。
func (s *EssayService) callLLMReview(userID uint, modelID uint, question models.EssayQuestion, materialTexts []string, answerTexts []string, userAnswer string) (*LLMReviewResult, error) {
	model, provider, err := s.getModelProvider(userID, modelID)
	if err != nil {
		return nil, fmt.Errorf("获取模型配置失败: %w", err)
	}
	if !model.Enabled || !provider.Enabled {
		return nil, errors.New("批改模型或服务商已禁用")
	}

	systemPrompt := buildReviewSystemPrompt()
	userPrompt := buildReviewUserPrompt(question, materialTexts, answerTexts, userAnswer)

	content, err := callOpenAICompatibleChat(provider.BaseURL, provider.APIKey, model.Name, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	content = extractJSONContent(content)
	var result LLMReviewResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析 LLM 批改结果失败: %w\nraw: %s", err, truncateForError(content, 500))
	}

	// 校验分数范围
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > float64(question.MaxScore) {
		result.Score = float64(question.MaxScore)
	}

	return &result, nil
}

func buildReviewSystemPrompt() string {
	return `你是一位资深的公务员考试申论阅卷专家。你需要根据题目要求、给定材料和参考答案/评分标准，对考生答案进行专业、客观、公正的批改评分。

## 评分原则
1. 严格对照参考答案中的评分要点进行逐点给分
2. 内容要点是评分核心，结构和语言是辅助项
3. 要点得分采用"踩点给分"原则：答到关键词或同义表达即给分
4. 总分不超过题目满分，各维度分之和应等于总分

## 输出要求
严格输出 JSON，不要输出 Markdown 代码块或其他格式。`
}

func buildReviewUserPrompt(question models.EssayQuestion, materialTexts []string, answerTexts []string, userAnswer string) string {
	var sb strings.Builder

	// ── 题目信息 ──
	sb.WriteString("# 批改任务\n\n")
	sb.WriteString("## 题目信息\n")
	sb.WriteString(fmt.Sprintf("- 题型: %s\n", question.QuestionType))
	sb.WriteString(fmt.Sprintf("- 满分: %d 分\n", question.MaxScore))
	sb.WriteString(fmt.Sprintf("- 字数限制: %d 字\n\n", question.WordLimit))
	if strings.TrimSpace(question.ScoringRubric) != "" {
		sb.WriteString("### 单题评分细则\n")
		sb.WriteString(question.ScoringRubric)
		sb.WriteString("\n\n")
	}
	sb.WriteString("### 题目原文\n")
	sb.WriteString(question.QuestionText)
	sb.WriteString("\n\n")

	// ── 给定材料（截断避免 token 溢出） ──
	if len(materialTexts) > 0 {
		sb.WriteString("## 给定材料\n")
		for i, mat := range materialTexts {
			sb.WriteString(fmt.Sprintf("### 材料 %d\n", i+1))
			runes := []rune(mat)
			if len(runes) > 3000 {
				sb.WriteString(string(runes[:2000]))
				sb.WriteString("\n...[材料过长已截断]...\n")
				sb.WriteString(string(runes[len(runes)-800:]))
			} else {
				sb.WriteString(mat)
			}
			sb.WriteString("\n\n")
		}
	}

	// ── 参考答案/评分标准 ──
	if len(answerTexts) > 0 {
		sb.WriteString("## 参考答案与评分标准\n")
		for i, ans := range answerTexts {
			sb.WriteString(fmt.Sprintf("### 参考答案 %d\n", i+1))
			sb.WriteString(ans)
			sb.WriteString("\n\n")
		}
	}

	// ── 考生答案 ──
	sb.WriteString("## 考生答案\n")
	sb.WriteString(userAnswer)
	sb.WriteString("\n\n")

	// ── 输出格式 ──
	sb.WriteString("## 请输出 JSON 格式的批改结果\n\n")
	sb.WriteString("```\n")
	sb.WriteString(`{
  "score": 得分(数字),
  "summary": "总体评价(100字以内)",
  "dimensions": [
    {"name": "内容要点", "score": 分数, "max_score": 满分, "comment": "具体点评"},
    {"name": "逻辑结构", "score": 分数, "max_score": 满分, "comment": "具体点评"},
    {"name": "语言表达", "score": 分数, "max_score": 满分, "comment": "具体点评"}
  ],
  "scoring_detail": "逐要点对照评分的详细说明",
  "highlights": [
    {"type": "good", "text": "答案中的亮点摘录", "comment": "为什么好"},
    {"type": "issue", "text": "答案中的问题摘录", "comment": "问题所在和改进建议"}
  ],
  "suggestions": ["改进建议1", "改进建议2"]
}
`)
	sb.WriteString("```\n")

	return sb.String()
}

func (s *EssayService) DeleteDocument(userID uint, documentID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssaySectionRelation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		var questionIDs []uint
		if err := tx.Model(&models.EssayQuestion{}).Where("user_id = ? AND document_id = ?", userID, documentID).Pluck("id", &questionIDs).Error; err != nil {
			return err
		}
		if len(questionIDs) > 0 {
			if err := tx.Where("user_id = ? AND question_id IN ?", userID, questionIDs).Delete(&models.EssayReview{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssaySection{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayChunk{}).Error; err != nil {
			return err
		}
		result := tx.Where("user_id = ? AND id = ?", userID, documentID).Delete(&models.EssayDocument{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *EssayService) getDocument(userID uint, documentID uint) (*models.EssayDocument, error) {
	var document models.EssayDocument
	if err := s.db.Where("user_id = ? AND id = ?", userID, documentID).First(&document).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

// ============= 以下为兼容/辅助函数 =============

func normalizeEssayQuestionSections(sections []models.EssaySection) []models.EssaySection {
	questionStartRe := regexp.MustCompile(`(?m)(?:^|\n)\s*((?:第[一二三四五六七八九十\d]+题)|(?:[一二三四五六七八九十]+、)|(?:\d+[.．、]))`)
	normalized := make([]models.EssaySection, 0, len(sections))
	for _, sec := range sections {
		if sec.SectionType != string(parser.SectionQuestion) {
			normalized = append(normalized, sec)
			continue
		}
		matches := questionStartRe.FindAllStringSubmatchIndex(sec.Content, -1)
		if len(matches) <= 1 {
			if strings.TrimSpace(sec.QuestionNo) == "" {
				sec.QuestionNo = fmt.Sprintf("%d", countQuestionSections(normalized)+1)
			}
			normalized = append(normalized, sec)
			continue
		}
		for i, match := range matches {
			start := match[0]
			if i > 0 && start < len(sec.Content) && sec.Content[start] == '\n' {
				start++
			}
			end := len(sec.Content)
			if i+1 < len(matches) {
				end = matches[i+1][0]
			}
			content := strings.TrimSpace(sec.Content[start:end])
			if content == "" {
				continue
			}
			next := sec
			next.ID = 0
			next.Content = content
			next.Title = ""
			next.QuestionNo = questionNoFromMarker(sec.Content[match[2]:match[3]])
			if next.QuestionNo == "" {
				next.QuestionNo = fmt.Sprintf("%d", countQuestionSections(normalized)+1)
			}
			normalized = append(normalized, next)
		}
	}
	return normalized
}

func countQuestionSections(sections []models.EssaySection) int {
	count := 0
	for _, sec := range sections {
		if sec.SectionType == string(parser.SectionQuestion) {
			count++
		}
	}
	return count
}

func questionNoFromMarker(marker string) string {
	marker = strings.TrimSpace(marker)
	marker = strings.TrimSuffix(marker, "题")
	marker = strings.TrimRight(marker, "、.．")
	marker = strings.TrimPrefix(marker, "第")
	if marker == "" {
		return ""
	}
	if regexp.MustCompile(`^\d+$`).MatchString(marker) {
		return marker
	}
	if value := chineseNumber(marker); value > 0 {
		return fmt.Sprintf("%d", value)
	}
	return marker
}

func chineseNumber(value string) int {
	digits := map[rune]int{'一': 1, '二': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9}
	if value == "十" {
		return 10
	}
	runes := []rune(value)
	if len(runes) == 1 {
		return digits[runes[0]]
	}
	if len(runes) == 2 && runes[0] == '十' {
		return 10 + digits[runes[1]]
	}
	if len(runes) == 2 && runes[1] == '十' {
		return digits[runes[0]] * 10
	}
	if len(runes) == 3 && runes[1] == '十' {
		return digits[runes[0]]*10 + digits[runes[2]]
	}
	return 0
}

func questionTitle(documentTitle string, content string, questionNo string, index int) string {
	line := strings.TrimSpace(strings.Split(content, "\n")[0])
	runes := []rune(line)
	if len(runes) > 24 {
		line = string(runes[:24]) + "..."
	}
	if strings.TrimSpace(questionNo) == "" {
		questionNo = fmt.Sprintf("%d", index)
	}
	prefix := strings.TrimSpace(documentTitle)
	if prefix == "" {
		prefix = "申论"
	}
	if line == "" {
		return fmt.Sprintf("%s - 第%s题", prefix, questionNo)
	}
	return fmt.Sprintf("%s - 第%s题 - %s", prefix, questionNo, line)
}

func guessQuestionType(content string) string {
	for _, item := range []string{"归纳概括", "综合分析", "提出对策", "应用文写作", "文章论述", "公文写作"} {
		if strings.Contains(content, item) {
			return item + "题"
		}
	}
	if strings.Contains(content, "概括") {
		return "归纳概括题"
	}
	if strings.Contains(content, "对策") || strings.Contains(content, "建议") {
		return "提出对策题"
	}
	if strings.Contains(content, "文章") || strings.Contains(content, "议论文") {
		return "文章论述题"
	}
	if strings.Contains(content, "分析") {
		return "综合分析题"
	}
	return "待确认"
}

func guessMaxScore(content string) int {
	re := regexp.MustCompile(`(\d+)\s*分`)
	matches := re.FindStringSubmatch(content)
	if len(matches) == 2 {
		var score int
		_, _ = fmt.Sscanf(matches[1], "%d", &score)
		if score > 0 && score <= 100 {
			return score
		}
	}
	return 100
}

func guessWordLimit(content string) int {
	re := regexp.MustCompile(`(\d+)\s*字`)
	matches := re.FindStringSubmatch(content)
	if len(matches) == 2 {
		var limit int
		_, _ = fmt.Sscanf(matches[1], "%d", &limit)
		if limit > 0 {
			return limit
		}
	}
	return 500
}

func mockEssayScore(answer string, wordLimit int) float64 {
	count := len([]rune(strings.ReplaceAll(answer, " ", "")))
	score := 68.0
	if wordLimit > 0 {
		ratio := float64(count) / float64(wordLimit)
		if ratio >= 0.75 && ratio <= 1.2 {
			score += 8
		}
	}
	if strings.Contains(answer, "首先") || strings.Contains(answer, "其次") {
		score += 4
	}
	if strings.Contains(answer, "材料") {
		score += 3
	}
	if score > 92 {
		return 92
	}
	return score
}

func truncateNote(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > 900 {
		return string(runes[:900])
	}
	return string(runes)
}

func maxSectionPageEnd(sections []models.EssaySection) int {
	maxValue := 0
	for _, sec := range sections {
		if sec.PageEnd > maxValue {
			maxValue = sec.PageEnd
		}
	}
	return maxValue
}

func sampleEssayPDFText(title string) string {
	if strings.TrimSpace(title) == "" {
		title = "申论 PDF"
	}
	parts := []string{
		fmt.Sprintf("材料一\n%s 围绕基层治理、公共服务和群众诉求展开。部分地区在信息收集、责任落实和反馈机制上仍有短板。", title),
		"材料二\n有群众反映办事流程较长，部门之间协同不足。基层工作人员表示，事项多、标准不统一，影响了办理效率。",
		"第一题\n请根据给定资料，概括基层治理中存在的主要问题。要求：全面、准确、有条理，不超过300字。20分",
		"第二题\n请结合给定资料，提出提升公共服务效率的对策建议。要求：措施具体、可操作，不超过500字。30分",
		"参考答案\n第一题参考要点：信息收集不充分、部门协同不足、责任边界不清、反馈机制不完善。第二题参考要点：建立统一平台、压实部门责任、优化流程、完善监督反馈。",
		"评分规则\n概括题每个要点4分，逻辑条理4分。对策题每条有效对策5分，针对性和可行性10分。",
	}
	sort.Strings(parts[:2])
	return strings.Join(parts, "\n\n")
}
