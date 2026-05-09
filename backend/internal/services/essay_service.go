package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"

	"gkweb/backend/internal/models"
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

func (s *EssayService) ParseDocument(userID uint, documentID uint, rawText string) (*models.EssayDocument, []models.EssayChunk, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, nil, err
	}
	document.Status = "parsing"
	document.Note = ""
	_ = s.db.Save(document).Error

	if strings.TrimSpace(rawText) == "" {
		if strings.TrimSpace(document.FilePath) != "" {
			pages, err := ExtractPDFTextPages(document.FilePath)
			if err != nil {
				return nil, nil, err
			}
			rawText = pagesToEssayText(pages)
		} else {
			rawText = sampleEssayPDFText(document.Title)
		}
	}

	chunks := splitEssayText(userID, document.ID, rawText)
	if len(chunks) == 0 {
		return nil, nil, errors.New("no chunks parsed from document")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND document_id = ?", userID, document.ID).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, document.ID).Delete(&models.EssayQuestion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, document.ID).Delete(&models.EssayChunk{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&chunks).Error; err != nil {
			return err
		}
		document.Status = "parsed"
		document.ChunkCount = len(chunks)
		document.PageCount = maxPageNo(chunks)
		document.Note = ""
		return tx.Save(document).Error
	})
	if err != nil {
		return nil, nil, err
	}

	return document, chunks, nil
}

func (s *EssayService) ListChunks(userID uint, documentID uint) ([]models.EssayChunk, error) {
	var chunks []models.EssayChunk
	err := s.db.Where("user_id = ? AND document_id = ?", userID, documentID).Order("chunk_index asc").Find(&chunks).Error
	return chunks, err
}

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
		return nil, errors.New("document has no chunks, parse it first")
	}

	for index := range chunks {
		chunkType, confidence, note := classifyEssayChunk(chunks[index].Content, document.DocumentRole)
		chunks[index].ChunkType = chunkType
		chunks[index].Confidence = confidence
		chunks[index].ClassModelID = modelID
		chunks[index].ClassificationNote = note
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for index := range chunks {
			if err := tx.Save(&chunks[index]).Error; err != nil {
				return err
			}
		}
		document.Status = "classified"
		document.ClassModelID = modelID
		return tx.Save(document).Error
	})
	return chunks, err
}

func (s *EssayService) AssembleQuestions(userID uint, documentID uint) ([]models.EssayQuestion, error) {
	document, err := s.getDocument(userID, documentID)
	if err != nil {
		return nil, err
	}

	chunks, err := s.ListChunks(userID, documentID)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, errors.New("document has no chunks")
	}

	questions := make([]models.EssayQuestion, 0)
	relations := make([]models.EssayQuestionChunk, 0)
	materialChunkIDs := make([]uint, 0)
	referenceChunkIDs := make([]uint, 0)
	scoringChunkIDs := make([]uint, 0)

	for _, chunk := range chunks {
		switch chunk.ChunkType {
		case EssayChunkMaterial:
			materialChunkIDs = append(materialChunkIDs, chunk.ID)
		case EssayChunkReferenceAnswer:
			referenceChunkIDs = append(referenceChunkIDs, chunk.ID)
		case EssayChunkScoringRule:
			scoringChunkIDs = append(scoringChunkIDs, chunk.ID)
		case EssayChunkQuestion:
			question := models.EssayQuestion{
				BaseModel:    models.BaseModel{UserID: userID},
				DocumentID:   documentID,
				Title:        questionTitle(chunk.Content, len(questions)+1),
				QuestionType: guessQuestionType(chunk.Content),
				QuestionText: chunk.Content,
				MaxScore:     guessMaxScore(chunk.Content),
				WordLimit:    guessWordLimit(chunk.Content),
				Status:       "assembled",
			}
			questions = append(questions, question)
		}
	}

	if len(questions) == 0 {
		return nil, errors.New("no question chunks found, classify document first")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestionChunk{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND document_id = ?", userID, documentID).Delete(&models.EssayQuestion{}).Error; err != nil {
			return err
		}
		for index := range questions {
			if err := tx.Create(&questions[index]).Error; err != nil {
				return err
			}
			for _, chunkID := range materialChunkIDs {
				relations = append(relations, relation(userID, documentID, questions[index].ID, chunkID, EssayChunkMaterial))
			}
			for _, chunkID := range referenceChunkIDs {
				relations = append(relations, relation(userID, documentID, questions[index].ID, chunkID, EssayChunkReferenceAnswer))
			}
			for _, chunkID := range scoringChunkIDs {
				relations = append(relations, relation(userID, documentID, questions[index].ID, chunkID, EssayChunkScoringRule))
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

func (s *EssayService) ListQuestions(userID uint, documentID uint) ([]models.EssayQuestion, error) {
	var questions []models.EssayQuestion
	err := s.db.Where("user_id = ? AND document_id = ?", userID, documentID).Order("id asc").Find(&questions).Error
	return questions, err
}

func (s *EssayService) ReviewAnswer(userID uint, questionID uint, modelID uint, userAnswer string) (*EssayReviewResult, error) {
	var question models.EssayQuestion
	if err := s.db.Where("user_id = ? AND id = ?", userID, questionID).First(&question).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(userAnswer) == "" {
		return nil, errors.New("user_answer is required")
	}

	score := mockEssayScore(userAnswer, question.WordLimit)
	resultPayload := map[string]any{
		"summary": "当前为流程骨架批改结果：已按题目、材料、参考答案和评分规则的结构化关系预留高质量模型调用位置。",
		"suggestions": []string{
			"后续接入高质量模型后，将把关联材料 chunk、参考答案 chunk 和评分规则 chunk 一并传入。",
			"建议先检查题目组装是否正确，再提交答案批改。",
		},
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

	return &EssayReviewResult{
		Review:      review,
		Score:       score,
		MaxScore:    question.MaxScore,
		Summary:     resultPayload["summary"].(string),
		Suggestions: resultPayload["suggestions"].([]string),
	}, nil
}

func (s *EssayService) getDocument(userID uint, documentID uint) (*models.EssayDocument, error) {
	var document models.EssayDocument
	if err := s.db.Where("user_id = ? AND id = ?", userID, documentID).First(&document).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

func splitEssayText(userID uint, documentID uint, rawText string) []models.EssayChunk {
	rawText = sanitizePostgresText(rawText)
	pageParts := splitPages(rawText)
	chunks := make([]models.EssayChunk, 0)
	chunkIndex := 1
	for pageIndex, pageText := range pageParts {
		for _, content := range splitParagraphChunks(pageText) {
			chunks = append(chunks, models.EssayChunk{
				BaseModel:  models.BaseModel{UserID: userID},
				DocumentID: documentID,
				PageNo:     pageIndex + 1,
				ChunkIndex: chunkIndex,
				Content:    content,
				ChunkType:  EssayChunkUnknown,
			})
			chunkIndex++
		}
	}
	return chunks
}

func splitPages(rawText string) []string {
	normalized := strings.ReplaceAll(rawText, "\r\n", "\n")
	re := regexp.MustCompile(`(?m)^-{3,}\s*page\s+\d+\s*-{3,}$`)
	parts := re.Split(normalized, -1)
	if len(parts) == 1 {
		parts = strings.Split(normalized, "\f")
	}
	return nonEmptyStrings(parts)
}

func splitParagraphChunks(pageText string) []string {
	parts := regexp.MustCompile(`\n\s*\n+`).Split(pageText, -1)
	result := make([]string, 0)
	for _, part := range parts {
		text := strings.TrimSpace(sanitizePostgresText(part))
		if text == "" {
			continue
		}
		if len([]rune(text)) <= 900 {
			result = append(result, text)
			continue
		}
		result = append(result, splitLongText(text, 700)...)
	}
	return result
}

func splitLongText(text string, size int) []string {
	runes := []rune(text)
	result := make([]string, 0)
	for start := 0; start < len(runes); start += size {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		result = append(result, strings.TrimSpace(string(runes[start:end])))
	}
	return result
}

func classifyEssayChunk(content string, documentRole string) (string, float64, string) {
	text := strings.ToLower(content)
	switch {
	case strings.Contains(content, "参考答案") || strings.Contains(content, "参考要点") || strings.Contains(content, "答案要点"):
		return EssayChunkReferenceAnswer, 0.82, "命中参考答案关键词"
	case strings.Contains(content, "评分") || strings.Contains(content, "赋分") || strings.Contains(content, "每点") || strings.Contains(content, "分）"):
		return EssayChunkScoringRule, 0.76, "命中评分规则关键词"
	case strings.Contains(content, "请根据") || strings.Contains(content, "请结合") || strings.Contains(content, "作答要求") || strings.Contains(content, "要求："):
		return EssayChunkQuestion, 0.8, "命中题目/作答要求关键词"
	case strings.Contains(content, "解析") || strings.Contains(content, "思路") || strings.Contains(text, "explanation"):
		return EssayChunkExplanation, 0.7, "命中解析说明关键词"
	case documentRole == "answer_key":
		return EssayChunkReferenceAnswer, 0.62, "文档类型为答案卷，未命中评分关键词时默认归为参考答案"
	case documentRole == "explanation":
		return EssayChunkExplanation, 0.62, "文档类型为解析卷，默认归为解析说明"
	case documentRole == "question_paper" && looksLikeQuestion(content):
		return EssayChunkQuestion, 0.66, "文档类型为题目卷，并命中题目句式"
	case documentRole == "question_paper":
		return EssayChunkMaterial, 0.6, "文档类型为题目卷，未命中题目句式时默认归为材料"
	case strings.Contains(content, "材料") || strings.Contains(content, "资料") || len([]rune(content)) > 180:
		return EssayChunkMaterial, 0.68, "按材料关键词或长段落归类"
	default:
		return EssayChunkUnknown, 0.4, "未命中稳定规则，等待 LLM 分类"
	}
}

func looksLikeQuestion(content string) bool {
	return strings.Contains(content, "请") ||
		strings.Contains(content, "要求") ||
		strings.Contains(content, "概括") ||
		strings.Contains(content, "分析") ||
		strings.Contains(content, "提出") ||
		strings.Contains(content, "不超过") ||
		strings.Contains(content, "字")
}

func relation(userID uint, documentID uint, questionID uint, chunkID uint, relationType string) models.EssayQuestionChunk {
	return models.EssayQuestionChunk{
		BaseModel:    models.BaseModel{UserID: userID},
		DocumentID:   documentID,
		QuestionID:   questionID,
		ChunkID:      chunkID,
		RelationType: relationType,
	}
}

func questionTitle(content string, index int) string {
	line := strings.TrimSpace(strings.Split(content, "\n")[0])
	runes := []rune(line)
	if len(runes) > 28 {
		line = string(runes[:28]) + "..."
	}
	if line == "" {
		return fmt.Sprintf("申论题目 %d", index)
	}
	return line
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

func nonEmptyStrings(parts []string) []string {
	result := make([]string, 0)
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func maxPageNo(chunks []models.EssayChunk) int {
	maxValue := 0
	for _, chunk := range chunks {
		if chunk.PageNo > maxValue {
			maxValue = chunk.PageNo
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
