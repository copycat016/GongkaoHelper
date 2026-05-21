package parser

import (
	"fmt"
	"strings"
)

// CleanedDocument 是 OCR/PDF 原文清洗后的中间结果。
// 保留 page/block/line 供规则引擎兜底使用。
type CleanedDocument struct {
	DocumentID string      `json:"document_id"`
	Blocks     []TextBlock `json:"blocks"`
}

// ────────────────────────────────────────────────────────
// 新流程：LLM 直接按行号范围切分整份文本
// ────────────────────────────────────────────────────────

// NumberedLine 是带行号的一行清洗后的文本。
type NumberedLine struct {
	LineNo int    `json:"line_no"`
	Text   string `json:"text"`
}

// BoundaryPlan 是 LLM 的结构化输出。
type BoundaryPlan struct {
	PaperTitle string            `json:"paper_title"`
	Sections   []BoundarySection `json:"sections"`
}

// BoundarySection 表示 LLM 识别的一个语义区域。
// 新流程使用 StartLine/EndLine（行号），旧流程使用 StartBlockID/EndBlockID。
type BoundarySection struct {
	SectionType        string   `json:"section_type"`
	Title              string   `json:"title"`
	StartLine          int      `json:"start_line"`                     // 新流程：起始行号（含）
	EndLine            int      `json:"end_line"`                       // 新流程：结束行号（含）
	StartBlockID       string   `json:"start_block_id,omitempty"`       // 旧流程兼容
	EndBlockID         string   `json:"end_block_id,omitempty"`         // 旧流程兼容
	QuestionNo         string   `json:"question_no,omitempty"`          // 题号（question 必填）
	RelatedQuestionNos []string `json:"related_question_nos,omitempty"` // 关联题号（answer 用）
	MaterialNos        []string `json:"material_nos,omitempty"`         // 引用材料编号（question 用）
	Confidence         float64  `json:"confidence"`
	Reason             string   `json:"reason"`
}

// PrepareNumberedLines 将原始文本清洗后生成带行号的文本行。
// 只做最基本的清洗：去空行、去页码、去分页标记、trim。
func PrepareNumberedLines(rawText string) []NumberedLine {
	rawText = strings.ReplaceAll(rawText, "\r\n", "\n")
	rawText = strings.ReplaceAll(rawText, "\r", "\n")
	raw := strings.Split(rawText, "\n")

	var lines []NumberedLine
	lineNo := 1
	for _, line := range raw {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if isPageMarker(strings.ToLower(trimmed)) {
			continue
		}
		if pageNumberRegex.MatchString(trimmed) {
			continue
		}
		lines = append(lines, NumberedLine{LineNo: lineNo, Text: trimmed})
		lineNo++
	}
	return lines
}

// BuildBoundaryPromptFromLines 直接将带行号的文本交给 LLM，让其按行号范围切分。
func BuildBoundaryPromptFromLines(lines []NumberedLine) string {
	var sb strings.Builder

	sb.WriteString("# 任务：申论试卷结构化切分\n\n")
	sb.WriteString("请将下面的文档按行号范围切分为不同区域。\n\n")

	// ── 结构知识 ──
	sb.WriteString("## 申论试卷结构\n\n")
	sb.WriteString("一份申论试卷通常依次包含：\n")
	sb.WriteString("1. **材料区 (material)**：「给定资料一」「给定资料二」…… 占文档大部分\n")
	sb.WriteString("2. **题目区 (question)**：1-5 道题，注明分值、字数限制，有「请根据/结合材料」\n")
	sb.WriteString("3. **参考答案区 (answer)**：按题号给出参考答案和/或评分标准（可能没有）\n")
	sb.WriteString("4. **解析区 (analysis)**：命题意图、考查能力等（可选）\n\n")
	sb.WriteString("注意：\n")
	sb.WriteString("- 有些试卷题目穿插在材料之间\n")
	sb.WriteString("- 参考答案区中的「第一题：」「第二题：」是答案子段，不是 question\n")
	sb.WriteString("- 评分标准/评分细则属于 answer 类型\n\n")

	// ── 输出格式 ──
	sb.WriteString("## 输出要求\n\n")
	sb.WriteString("严格输出 JSON，不要输出 Markdown 代码块、不要输出解释文字。\n\n")
	sb.WriteString(`{
  "paper_title": "试卷标题或空字符串",
  "sections": [
    {
      "section_type": "material|question|answer|analysis|unknown",
      "title": "小标题",
      "start_line": 起始行号,
      "end_line": 结束行号,
      "question_no": "题号(question必填)",
      "related_question_nos": ["answer关联的题号"],
      "material_nos": ["question引用的材料编号"],
      "confidence": 0.9,
      "reason": "判定依据"
    }
  ]
}`)
	sb.WriteString("\n\n")

	// ── 规则 ──
	sb.WriteString("## 规则\n\n")
	sb.WriteString("1. 行号范围不能重叠，必须覆盖第 1 行到第 ")
	sb.WriteString(fmt.Sprintf("%d", len(lines)))
	sb.WriteString(" 行\n")
	sb.WriteString("2. start_line / end_line 使用文档中每行开头 [LN] 中的 N\n")
	sb.WriteString("3. answer 区中出现的题号属于 answer，不要标为 question\n")
	sb.WriteString("4. 每个 question 必须填 question_no；每个 answer 必须填 related_question_nos\n\n")

	// ── 文档正文 ──
	sb.WriteString("## 文档正文（共 ")
	sb.WriteString(fmt.Sprintf("%d", len(lines)))
	sb.WriteString(" 行）\n\n")

	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("[L%d] %s\n", line.LineNo, line.Text))
	}

	return sb.String()
}

// BuildBoundarySystemPrompt 返回边界识别的 system prompt。
func BuildBoundarySystemPrompt() string {
	return "你是申论试卷结构化分析工具。严格输出 JSON，不输出 Markdown 代码块。你只识别边界和关联关系，绝不改写原文。"
}

// ApplyBoundaryPlanToLines 根据 LLM 的行号范围切分结果，从原始行中提取 section。
func ApplyBoundaryPlanToLines(lines []NumberedLine, plan BoundaryPlan) ([]Section, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("no lines")
	}
	if len(plan.Sections) == 0 {
		return nil, fmt.Errorf("boundary plan has no sections")
	}

	// 行号 → 索引
	lineIndex := make(map[int]int, len(lines))
	for i, line := range lines {
		lineIndex[line.LineNo] = i
	}

	sections := make([]Section, 0, len(plan.Sections))
	var skipped int

	for _, item := range plan.Sections {
		startIdx, ok1 := lineIndex[item.StartLine]
		endIdx, ok2 := lineIndex[item.EndLine]

		// 容错：行号不精确时就近匹配
		if !ok1 {
			startIdx, ok1 = findNearestLineIndex(lines, item.StartLine)
		}
		if !ok2 {
			endIdx, ok2 = findNearestLineIndex(lines, item.EndLine)
		}

		if !ok1 || !ok2 || endIdx < startIdx {
			skipped++
			continue
		}

		var textParts []string
		for i := startIdx; i <= endIdx; i++ {
			textParts = append(textParts, lines[i].Text)
		}
		text := strings.Join(textParts, "\n")

		secType := SectionType(normalizeSectionType(item.SectionType))
		if secType == "" {
			secType = SectionUnknown
		}

		conf := item.Confidence
		if conf <= 0 || conf > 1 {
			conf = 0.7
		}

		section := Section{
			Type:               secType,
			Title:              item.Title,
			Text:               text,
			PageStart:          1,
			PageEnd:            1,
			Confidence:         conf,
			Reason:             item.Reason,
			QuestionNo:         strings.TrimSpace(item.QuestionNo),
			RelatedQuestionNos: item.RelatedQuestionNos,
			MaterialNos:        item.MaterialNos,
		}

		if strings.TrimSpace(section.Title) == "" && len(textParts) > 0 {
			title := strings.TrimSpace(textParts[0])
			runes := []rune(title)
			if len(runes) > 40 {
				title = string(runes[:40]) + "..."
			}
			section.Title = title
		}

		sections = append(sections, section)
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("all %d sections had invalid line ranges", skipped)
	}

	counters := make(map[SectionType]int)
	for i := range sections {
		counters[sections[i].Type]++
		sections[i].ID = fmt.Sprintf("sec_%s_%03d", sections[i].Type, counters[sections[i].Type])
	}

	return sections, nil
}

func findNearestLineIndex(lines []NumberedLine, targetLineNo int) (int, bool) {
	if len(lines) == 0 {
		return 0, false
	}
	if targetLineNo <= lines[0].LineNo {
		return 0, true
	}
	if targetLineNo >= lines[len(lines)-1].LineNo {
		return len(lines) - 1, true
	}
	bestIdx := 0
	bestDist := intAbs(lines[0].LineNo - targetLineNo)
	for i, line := range lines {
		dist := intAbs(line.LineNo - targetLineNo)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}
	return bestIdx, true
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ────────────────────────────────────────────────────────
// 旧流程兼容：block-based（规则引擎 + debug 保留）
// ────────────────────────────────────────────────────────

// BuildBoundaryPrompt 旧版 block-based prompt。
func BuildBoundaryPrompt(blocks []TextBlock) string {
	var sb strings.Builder
	sb.WriteString("你是申论试卷结构化助手。你只能识别边界，不要改写、总结或补全文本。\n")
	sb.WriteString("请根据 block 顺序，把整份申论文档分成：material、question、answer、analysis、unknown。\n")
	sb.WriteString("请输出严格 JSON。\n")
	sb.WriteString(`{"paper_title":"","sections":[{"section_type":"","title":"","start_block_id":"","end_block_id":"","question_no":"","confidence":0.0,"reason":""}]}`)
	sb.WriteString("\n\n可用 block：\n")
	for _, block := range blocks {
		text := strings.TrimSpace(block.Text)
		if len([]rune(text)) > 600 {
			runes := []rune(text)
			text = string(runes[:400]) + "...[截断]..." + string(runes[len(runes)-150:])
		}
		sb.WriteString(fmt.Sprintf("[BLOCK:%s] page=%d\n%s\n\n", block.ID, block.PageNo, text))
	}
	return sb.String()
}

// ApplyBoundaryPlan 旧版 block-based apply。
func ApplyBoundaryPlan(blocks []TextBlock, plan BoundaryPlan) ([]Section, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks")
	}
	if len(plan.Sections) == 0 {
		return nil, fmt.Errorf("boundary plan has no sections")
	}

	blockIndex := make(map[string]int, len(blocks))
	for i, block := range blocks {
		blockIndex[block.ID] = i
	}

	sections := make([]Section, 0, len(plan.Sections))
	var skipped int
	for _, item := range plan.Sections {
		bid := item.StartBlockID
		eid := item.EndBlockID
		if bid == "" || eid == "" {
			skipped++
			continue
		}
		start, ok := blockIndex[bid]
		if !ok {
			skipped++
			continue
		}
		end, ok := blockIndex[eid]
		if !ok {
			skipped++
			continue
		}
		if end < start {
			skipped++
			continue
		}

		sectionBlocks := append([]TextBlock(nil), blocks[start:end+1]...)
		secType := SectionType(normalizeSectionType(item.SectionType))
		if secType == "" {
			secType = SectionUnknown
		}
		section := buildSectionFromBlocks(secType, item.Title, sectionBlocks)
		section.Confidence = item.Confidence
		if section.Confidence <= 0 || section.Confidence > 1 {
			section.Confidence = 0.7
		}
		section.Reason = item.Reason
		section.QuestionNo = strings.TrimSpace(item.QuestionNo)
		section.RelatedQuestionNos = item.RelatedQuestionNos
		section.MaterialNos = item.MaterialNos
		sections = append(sections, section)
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("all %d sections had invalid block IDs", skipped)
	}

	counters := make(map[SectionType]int)
	for i := range sections {
		counters[sections[i].Type]++
		sections[i].ID = fmt.Sprintf("sec_%s_%03d", sections[i].Type, counters[sections[i].Type])
		if strings.TrimSpace(sections[i].Title) == "" {
			sections[i].Title = extractTitle(sections[i])
		}
	}
	return sections, nil
}

func buildSectionFromBlocks(secType SectionType, title string, blocks []TextBlock) Section {
	pageStart := blocks[0].PageNo
	pageEnd := blocks[0].PageNo
	texts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if block.PageNo < pageStart {
			pageStart = block.PageNo
		}
		if block.PageNo > pageEnd {
			pageEnd = block.PageNo
		}
		texts = append(texts, block.Text)
	}
	return Section{
		Type:      secType,
		Title:     title,
		Blocks:    blocks,
		Text:      strings.Join(texts, "\n\n"),
		PageStart: pageStart,
		PageEnd:   pageEnd,
	}
}
