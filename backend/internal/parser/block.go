package parser

import (
	"fmt"
	"strings"
)

// TextLine 表示 OCR/PDF 提取的一行原始文本。
type TextLine struct {
	ID      string `json:"id"`
	BlockID string `json:"block_id,omitempty"`
	Text    string `json:"text"`
	PageNo  int    `json:"page_no"`
}

// TextBlock 是清洗后的基本单元，由连续的多行组成，保留原始位置信息。
type TextBlock struct {
	ID      string     `json:"id"`
	PageNo  int        `json:"page_no"`
	Lines   []TextLine `json:"lines"`
	RawText string     `json:"raw_text"` // 原始拼接文本（未合并断行）
	Text    string     `json:"text"`     // 清洗后的文本
	Meta    BlockMeta  `json:"meta"`
}

// BlockMeta 保存用于后续分析的元信息。
type BlockMeta struct {
	IsHeader     bool `json:"is_header"`
	IsFooter     bool `json:"is_footer"`
	IsPageNumber bool `json:"is_page_number"`
	IsEmpty      bool `json:"is_empty"`
	IsDuplicate  bool `json:"is_duplicate"`
	LineCount    int  `json:"line_count"`
	CharCount    int  `json:"char_count"`
}

// SectionType 表示文档结构区域类型。
type SectionType string

const (
	SectionUnknown  SectionType = "unknown"
	SectionMaterial SectionType = "material"
	SectionQuestion SectionType = "question"
	SectionAnswer   SectionType = "answer"
	SectionAnalysis SectionType = "analysis"
)

func (s SectionType) String() string { return string(s) }

// Section 是一个语义区域，由多个 TextBlock 组成。
type Section struct {
	ID                 string      `json:"id"`
	Type               SectionType `json:"type"`
	Title              string      `json:"title"` // 自动提取的标题（通常是第一个 block 的首行）
	Blocks             []TextBlock `json:"blocks"`
	Text               string      `json:"text"` // 合并后的完整文本
	PageStart          int         `json:"page_start"`
	PageEnd            int         `json:"page_end"`
	Confidence         float64     `json:"confidence"`
	Reason             string      `json:"reason"`
	IsDividerStart     bool        `json:"is_divider_start"`          // 该 section 是否由 divider 锚点触发
	QuestionNo         string      `json:"question_no,omitempty"`     // 题号（question 区域使用）
	RelatedQuestionNos []string    `json:"related_question_nos,omitempty"` // 关联题号（answer/material 使用）
	MaterialNos        []string    `json:"material_nos,omitempty"`    // 引用的材料编号（question 使用）
}

// BlockRef 用于在其他结构中引用 Block。
type BlockRef struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// BuildBlocks 将按 page/block/line 组织的原始文本构建为 []TextBlock。
// 如果 TextLine.BlockID 存在，则优先按 OCR/PDF 原始 block 边界分组；
// 否则按每页连续非空行分组，空行作为分隔。
func BuildBlocks(lines []TextLine) []TextBlock {
	var blocks []TextBlock
	var current []TextLine
	currentBlockID := ""

	flush := func() {
		if len(current) == 0 {
			return
		}
		blockID := currentBlockID
		if blockID == "" {
			blockID = fmt.Sprintf("b%04d", len(blocks))
		}
		block := TextBlock{
			ID:      blockID,
			PageNo:  current[0].PageNo,
			Lines:   append([]TextLine(nil), current...),
			RawText: joinLines(current),
		}
		blocks = append(blocks, block)
		current = current[:0]
		currentBlockID = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line.Text)
		if trimmed == "" {
			flush()
			continue
		}
		if len(current) > 0 && (current[len(current)-1].PageNo != line.PageNo || shouldStartNewBlock(currentBlockID, line.BlockID)) {
			flush()
		}
		if currentBlockID == "" && line.BlockID != "" {
			currentBlockID = line.BlockID
		}
		current = append(current, line)
	}
	flush()

	for i := range blocks {
		blocks[i].Text = blocks[i].RawText
		blocks[i].Meta.LineCount = len(blocks[i].Lines)
		blocks[i].Meta.CharCount = len([]rune(blocks[i].Text))
	}
	return blocks
}

func shouldStartNewBlock(currentBlockID string, nextBlockID string) bool {
	if currentBlockID == "" || nextBlockID == "" {
		return false
	}
	return currentBlockID != nextBlockID
}

func joinLines(lines []TextLine) string {
	parts := make([]string, len(lines))
	for i, l := range lines {
		parts[i] = l.Text
	}
	return strings.Join(parts, "\n")
}
