package parser

import (
	"fmt"
	"strings"
)

// StructuredResult 是 parser 的最终输出，可直接用于后续批改。
type StructuredResult struct {
	DocumentID string     `json:"document_id"`
	Sections   []Section  `json:"sections"`
	Questions  []Question `json:"questions"`
	Materials  []Material `json:"materials"`
	Answers    []Answer   `json:"answers"`
	Relations  []Relation `json:"relations"`
}

// Question 表示一道题目及其关联的材料和答案。
type Question struct {
	ID           string   `json:"id"`
	SectionID    string   `json:"section_id"`
	Title        string   `json:"title"`
	QuestionText string   `json:"question_text"`
	QuestionType string   `json:"question_type"`
	MaxScore     int      `json:"max_score"`
	WordLimit    int      `json:"word_limit"`
	QuestionNo   string   `json:"question_no"`            // 题号，如 "1","2"
	MaterialNos  []string `json:"material_nos,omitempty"` // 引用的材料编号
	MaterialIDs  []string `json:"material_ids"`
	AnswerIDs    []string `json:"answer_ids"`
}

// Material 表示一段材料。
type Material struct {
	ID        string `json:"id"`
	SectionID string `json:"section_id"`
	Title     string `json:"title"`
	Text      string `json:"text"`
}

// Answer 表示参考答案或评分标准。
type Answer struct {
	ID                 string   `json:"id"`
	SectionID          string   `json:"section_id"`
	Title              string   `json:"title"`
	Text               string   `json:"text"`
	RelatedQuestionNos []string `json:"related_question_nos,omitempty"` // 关联的题号
}

// Relation 表示 section 之间的关联。
type Relation struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
	Type   string `json:"type"` // e.g. "question_material", "question_answer"
}

// BuildResult 从 sections 构建结构化结果，建立 question-material-answer 关联。
//
// 关联策略（优先级从高到低）：
// 1. 精确关联：利用 BoundaryPlan 中 LLM 输出的 question_no / related_question_nos / material_nos
// 2. 编号推断：如果 answer section 没有 related_question_nos，尝试按同序号一一对应
// 3. 位置兜底：如果以上均无法确定，按位置关系做保守关联
func BuildResult(documentID string, sections []Section) StructuredResult {
	result := StructuredResult{DocumentID: documentID, Sections: sections}

	var materials []Material
	var answers []Answer
	var questions []Question
	questionCounter := 0

	// 收集材料编号索引（如 "材料一" → material_001 的 section ID）
	materialNoIndex := make(map[string]string) // material_no -> section.ID

	for _, sec := range sections {
		switch sec.Type {
		case SectionMaterial:
			m := Material{
				ID:        sec.ID,
				SectionID: sec.ID,
				Title:     sec.Title,
				Text:      sec.Text,
			}
			materials = append(materials, m)
			// 从 title 推断材料编号
			matNo := extractMaterialNo(sec.Title)
			if matNo != "" {
				materialNoIndex[matNo] = sec.ID
			}
		case SectionAnswer:
			answers = append(answers, Answer{
				ID:                 sec.ID,
				SectionID:          sec.ID,
				Title:              sec.Title,
				Text:               sec.Text,
				RelatedQuestionNos: sec.RelatedQuestionNos,
			})
		case SectionQuestion:
			questionCounter++
			qNo := sec.QuestionNo
			if qNo == "" {
				qNo = fmt.Sprintf("%d", questionCounter)
			}
			q := Question{
				ID:           fmt.Sprintf("q_%03d", questionCounter),
				SectionID:    sec.ID,
				Title:        sec.Title,
				QuestionText: sec.Text,
				QuestionType: guessQuestionType(sec.Text),
				MaxScore:     guessMaxScore(sec.Text),
				WordLimit:    guessWordLimit(sec.Text),
				QuestionNo:   qNo,
				MaterialNos:  sec.MaterialNos,
			}
			questions = append(questions, q)
		}
	}

	result.Materials = materials
	result.Answers = answers
	result.Questions = questions

	// ── 建立关联 ──

	// question_no -> Question index
	qNoIndex := make(map[string]int)
	for i, q := range questions {
		qNoIndex[q.QuestionNo] = i
	}

	// 第一步：利用 LLM 输出的 material_nos 精确关联材料
	for i, q := range questions {
		if len(q.MaterialNos) > 0 {
			for _, mNo := range q.MaterialNos {
				if secID, ok := materialNoIndex[mNo]; ok {
					q.MaterialIDs = appendUnique(q.MaterialIDs, secID)
					result.Relations = append(result.Relations, Relation{
						FromID: q.ID, ToID: secID, Type: "question_material",
					})
				}
			}
		}
		// 如果题目没有指定材料编号，关联所有材料（保守策略）
		if len(q.MaterialIDs) == 0 {
			for _, m := range materials {
				q.MaterialIDs = appendUnique(q.MaterialIDs, m.ID)
				result.Relations = append(result.Relations, Relation{
					FromID: q.ID, ToID: m.ID, Type: "question_material",
				})
			}
		}
		result.Questions[i] = q
	}

	// 第二步：利用 related_question_nos 精确关联答案
	answersLinked := make(map[int]bool) // answer index -> 是否已关联
	for aIdx, a := range answers {
		if len(a.RelatedQuestionNos) > 0 {
			for _, qNo := range a.RelatedQuestionNos {
				if qIdx, ok := qNoIndex[qNo]; ok {
					q := result.Questions[qIdx]
					q.AnswerIDs = appendUnique(q.AnswerIDs, a.ID)
					result.Relations = append(result.Relations, Relation{
						FromID: q.ID, ToID: a.ID, Type: "question_answer",
					})
					result.Questions[qIdx] = q
					answersLinked[aIdx] = true
				}
			}
		}
	}

	// 第三步：未关联的答案按顺序一一对应
	if len(questions) > 0 {
		unlinkedAnswers := make([]int, 0)
		for aIdx := range answers {
			if !answersLinked[aIdx] {
				unlinkedAnswers = append(unlinkedAnswers, aIdx)
			}
		}
		if len(unlinkedAnswers) > 0 {
			if len(unlinkedAnswers) == len(questions) {
				// 答案数量与题目数量相同，按顺序一一对应
				for i, aIdx := range unlinkedAnswers {
					a := answers[aIdx]
					q := result.Questions[i]
					q.AnswerIDs = appendUnique(q.AnswerIDs, a.ID)
					result.Relations = append(result.Relations, Relation{
						FromID: q.ID, ToID: a.ID, Type: "question_answer",
					})
					result.Questions[i] = q
				}
			} else {
				// 数量不匹配，所有未关联答案关联到所有题目（保守兜底）
				for _, aIdx := range unlinkedAnswers {
					a := answers[aIdx]
					for qIdx := range questions {
						q := result.Questions[qIdx]
						q.AnswerIDs = appendUnique(q.AnswerIDs, a.ID)
						result.Relations = append(result.Relations, Relation{
							FromID: q.ID, ToID: a.ID, Type: "question_answer",
						})
						result.Questions[qIdx] = q
					}
				}
			}
		}
	}

	return result
}

// extractMaterialNo 从标题中提取材料编号，如 "给定资料一" → "1"，"材料3" → "3"
func extractMaterialNo(title string) string {
	cnNumMap := map[rune]string{
		'一': "1", '二': "2", '三': "3", '四': "4", '五': "5",
		'六': "6", '七': "7", '八': "8", '九': "9", '十': "10",
	}
	for r, n := range cnNumMap {
		if strings.ContainsRune(title, r) {
			return n
		}
	}
	// 尝试提取阿拉伯数字
	for _, r := range title {
		if r >= '1' && r <= '9' {
			return string(r)
		}
	}
	return ""
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// ContextForReview 为指定题目生成批改所需的上下文（question + related materials + answer）。
func (r StructuredResult) ContextForReview(questionID string) (question Question, materials []Material, answers []Answer, found bool) {
	for _, q := range r.Questions {
		if q.ID != questionID {
			continue
		}
		found = true
		question = q
		for _, m := range r.Materials {
			for _, mid := range q.MaterialIDs {
				if m.ID == mid {
					materials = append(materials, m)
					break
				}
			}
		}
		for _, a := range r.Answers {
			for _, aid := range q.AnswerIDs {
				if a.ID == aid {
					answers = append(answers, a)
					break
				}
			}
		}
		return
	}
	return
}

func guessQuestionType(text string) string {
	for _, item := range []string{"归纳概括", "综合分析", "提出对策", "应用文写作", "文章论述", "公文写作"} {
		if strings.Contains(text, item) {
			return item + "题"
		}
	}
	if strings.Contains(text, "概括") {
		return "归纳概括题"
	}
	if strings.Contains(text, "对策") || strings.Contains(text, "建议") {
		return "提出对策题"
	}
	if strings.Contains(text, "文章") || strings.Contains(text, "议论文") {
		return "文章论述题"
	}
	if strings.Contains(text, "分析") {
		return "综合分析题"
	}
	return "待确认"
}

func guessMaxScore(text string) int {
	// 简单正则提取分值
	re := strings.NewReplacer(
		"(", "", ")", "", "（", "", "）", "",
	)
	clean := re.Replace(text)
	for _, pattern := range []string{"(%d+)分", "满分(%d+)", "共(%d+)分", "(%d+)\\s*分"} {
		_ = pattern
	}
	// 简化实现：搜索 "XX分"
	idx := strings.Index(clean, "分")
	if idx > 0 {
		for i := idx - 1; i >= 0; i-- {
			if clean[i] < '0' || clean[i] > '9' {
				if i+1 < idx {
					var score int
					fmt.Sscanf(clean[i+1:idx], "%d", &score)
					if score > 0 && score <= 100 {
						return score
					}
				}
				break
			}
		}
	}
	return 100
}

func guessWordLimit(text string) int {
	// 搜索 "XXX字" 或 "不超过XXX字"
	patterns := []string{"不超过", "字数", "限制", "约", "左右"}
	clean := text
	for _, p := range patterns {
		clean = strings.ReplaceAll(clean, p, "")
	}
	idx := strings.Index(clean, "字")
	if idx > 0 {
		for i := idx - 1; i >= 0; i-- {
			if clean[i] < '0' || clean[i] > '9' {
				if i+1 < idx {
					var limit int
					fmt.Sscanf(clean[i+1:idx], "%d", &limit)
					if limit > 0 && limit < 10000 {
						return limit
					}
				}
				break
			}
		}
	}
	return 500
}
