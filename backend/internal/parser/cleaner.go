package parser

import (
	"regexp"
	"strings"
)

func CollapseBlocks(blocks []TextBlock) []TextBlock {
	if len(blocks) <= 1 {
		return blocks
	}
	merged := TextBlock{
		ID:     blocks[0].ID,
		PageNo: blocks[0].PageNo,
	}
	texts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		merged.Lines = append(merged.Lines, b.Lines...)
		if b.Text != "" {
			texts = append(texts, b.Text)
		}
		if b.PageNo > merged.PageNo {
			merged.PageNo = b.PageNo
		}
	}
	merged.RawText = joinLines(merged.Lines)
	merged.Text = strings.Join(texts, "\n")
	merged.Meta.LineCount = len(merged.Lines)
	merged.Meta.CharCount = len([]rune(merged.Text))
	return []TextBlock{merged}
}

// Cleaner 对 TextBlock 进行基础清洗。
type Cleaner struct {
	// 可配置开关
	RemoveHeaders     bool
	RemoveFooters     bool
	RemovePageNumbers bool
	RemoveEmptyBlocks bool
	DeduplicateLines  bool
	MergeBrokenLines  bool

	// 内部状态
	headerPatterns map[string]int
	footerPatterns map[string]int
}

// NewCleaner 创建默认配置的清洗器。
func NewCleaner() *Cleaner {
	return &Cleaner{
		RemoveHeaders:     true,
		RemoveFooters:     true,
		RemovePageNumbers: true,
		RemoveEmptyBlocks: true,
		DeduplicateLines:  true,
		MergeBrokenLines:  true,
	}
}

// Clean 执行完整清洗管道。
func (c *Cleaner) Clean(blocks []TextBlock) []TextBlock {
	if len(blocks) == 0 {
		return blocks
	}

	// 1. 检测页眉页脚模式
	if c.RemoveHeaders || c.RemoveFooters {
		c.detectHeaderFooter(blocks)
	}

	// 2. 逐 block 处理
	var result []TextBlock
	for _, block := range blocks {
		if c.shouldSkipBlock(block) {
			continue
		}

		lines := c.cleanBlockLines(block)
		if len(lines) == 0 {
			continue
		}

		block.Lines = lines
		block.Text = joinLines(lines)
		block.Meta.CharCount = len([]rune(block.Text))
		block.Meta.LineCount = len(lines)
		result = append(result, block)
	}

	// 3. 合并断行（跨 block 的短行合并）
	if c.MergeBrokenLines {
		result = c.mergeBrokenLines(result)
	}

	return result
}

func (c *Cleaner) detectHeaderFooter(blocks []TextBlock) {
	c.headerPatterns = make(map[string]int)
	c.footerPatterns = make(map[string]int)

	pageBlocks := make(map[int][]TextBlock)
	for _, b := range blocks {
		pageBlocks[b.PageNo] = append(pageBlocks[b.PageNo], b)
	}

	for _, pbs := range pageBlocks {
		if len(pbs) == 0 {
			continue
		}
		// 页眉：前2个 block 的首行
		for i := 0; i < min(2, len(pbs)); i++ {
			if len(pbs[i].Lines) > 0 {
				key := normalizePattern(pbs[i].Lines[0].Text)
				if key != "" {
					c.headerPatterns[key]++
				}
			}
		}
		// 页脚：后2个 block 的末行
		for i := max(0, len(pbs)-2); i < len(pbs); i++ {
			lines := pbs[i].Lines
			if len(lines) > 0 {
				key := normalizePattern(lines[len(lines)-1].Text)
				if key != "" {
					c.footerPatterns[key]++
				}
			}
		}
	}
}

func (c *Cleaner) shouldSkipBlock(block TextBlock) bool {
	if c.RemoveEmptyBlocks && block.Meta.IsEmpty {
		return true
	}
	if c.RemovePageNumbers && block.Meta.IsPageNumber {
		return true
	}
	if c.RemoveHeaders && block.Meta.IsHeader {
		return true
	}
	if c.RemoveFooters && block.Meta.IsFooter {
		return true
	}
	return false
}

func (c *Cleaner) cleanBlockLines(block TextBlock) []TextLine {
	lines := make([]TextLine, 0, len(block.Lines))

	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Text)
		if text == "" {
			continue
		}

		// 页码检测
		if c.RemovePageNumbers && isPageNumber(text) {
			continue
		}

		// 页眉页脚过滤
		if c.RemoveHeaders && c.headerPatterns[normalizePattern(text)] > 1 {
			continue
		}
		if c.RemoveFooters && c.footerPatterns[normalizePattern(text)] > 1 {
			continue
		}

		lines = append(lines, TextLine{ID: line.ID, BlockID: line.BlockID, Text: text, PageNo: line.PageNo})
	}

	// 重复行去重（只去重紧邻的完全重复）
	if c.DeduplicateLines {
		lines = dedupConsecutive(lines)
	}

	return lines
}

func (c *Cleaner) mergeBrokenLines(blocks []TextBlock) []TextBlock {
	if len(blocks) == 0 {
		return blocks
	}

	var result []TextBlock
	current := blocks[0]

	flush := func() {
		if current.Text != "" {
			result = append(result, current)
		}
	}

	for i := 1; i < len(blocks); i++ {
		next := blocks[i]
		if shouldMerge(current, next) {
			current.Lines = append(current.Lines, next.Lines...)
			current.Text = current.Text + " " + next.Text
			current.Meta.LineCount = len(current.Lines)
			current.Meta.CharCount = len([]rune(current.Text))
			if next.PageNo > current.PageNo {
				current.PageNo = next.PageNo // 取最大页码
			}
		} else {
			flush()
			current = next
		}
	}
	flush()
	return result
}

// shouldMerge 判断两个 block 是否应该合并（断行修复）。
func shouldMerge(a, b TextBlock) bool {
	if a.PageNo != b.PageNo {
		return false // 跨页不合并
	}
	aText := strings.TrimSpace(a.Text)
	bText := strings.TrimSpace(b.Text)
	if aText == "" || bText == "" {
		return false
	}

	// 如果 b 包含明显的语义锚点（如"第一题"、"参考答案"），不合并，避免破坏结构边界
	if len(b.Lines) > 0 {
		anchors := CompileAnchors(DefaultAnchors())
		if match := BestAnchorForBlock(b, anchors); match != nil && match.Anchor.Priority >= 80 {
			return false
		}
	}

	// 如果 a 的末尾没有结束标点，且 b 的开头没有开始标点，则可能是断行
	aLastRune := lastRune(aText)
	bFirstRune := firstRune(bText)

	isEndingPunct := isChineseEndingPunct(aLastRune)
	isStartingPunct := isChineseStartingPunct(bFirstRune)

	if !isEndingPunct && !isStartingPunct {
		if len([]rune(aText)) < 200 {
			return true
		}
	}
	return false
}

var pageNumberRegex = regexp.MustCompile(`^\s*[-—]?\s*\d+\s*[-—]?\s*$|^\s*第\s*\d+\s*[页頁]\s*$`)

func isPageNumber(text string) bool {
	return pageNumberRegex.MatchString(text)
}

func normalizePattern(text string) string {
	text = strings.TrimSpace(text)
	// 去掉数字变体，用于统计重复模式
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(text, "#")
}

func dedupConsecutive(lines []TextLine) []TextLine {
	if len(lines) == 0 {
		return lines
	}
	result := []TextLine{lines[0]}
	for i := 1; i < len(lines); i++ {
		if lines[i].Text != lines[i-1].Text {
			result = append(result, lines[i])
		}
	}
	return result
}

func isChineseEndingPunct(r rune) bool {
	return strings.ContainsRune("。！？；：，、）】》」』", r)
}

func isChineseStartingPunct(r rune) bool {
	return strings.ContainsRune("（【《「『", r)
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func lastRune(s string) rune {
	var r rune
	for _, c := range s {
		r = c
	}
	return r
}
