package parser

import (
	"fmt"
	"strings"
)

// StateMachine 基于锚点将 block 序列切分为 section。
type StateMachine struct {
	anchors []Anchor
}

// NewStateMachine 创建状态机。
func NewStateMachine(anchors []Anchor) *StateMachine {
	return &StateMachine{anchors: CompileAnchors(anchors)}
}

// Run 执行切分，返回 sections。
// 策略：
// 1. 每个 block 检查是否有锚点。
// 2. 如果是分隔锚点（IsDivider=true），即使类型相同也创建新 section。
// 3. 如果是类型锚点（IsDivider=false），仅在类型变化时创建新 section。
// 4. 无锚点时，继承当前状态（Unknown 除外）。
func (sm *StateMachine) Run(blocks []TextBlock) []Section {
	if len(blocks) == 0 {
		return nil
	}

	var sections []Section
	currentType := SectionUnknown
	var currentBlocks []TextBlock
	var currentDividers []bool // 记录每个 block 是否由 divider 锚点触发

	flush := func() {
		if len(currentBlocks) == 0 {
			return
		}
		sec := sm.buildSection(currentType, currentBlocks, currentDividers)
		sections = append(sections, sec)
		currentBlocks = nil
		currentDividers = nil
	}

	for _, block := range blocks {
		match := BestAnchorForBlock(block, sm.anchors)

		if match != nil {
			newType := match.Anchor.Type
			isDivider := match.Anchor.IsDivider

			// 分隔锚点：总是创建新 section
			// 类型锚点：仅在类型变化时创建新 section
			if isDivider || newType != currentType {
				flush()
				currentType = newType
			}
			currentBlocks = append(currentBlocks, block)
			currentDividers = append(currentDividers, isDivider)
			continue
		}

		// 无锚点：如果是 Unknown，尝试猜测类型
		if currentType == SectionUnknown {
			guessed := GuessSectionTypeFromText(block.Text)
			if guessed != SectionUnknown && len(currentBlocks) > 0 {
				flush()
				currentType = guessed
			}
		}

		currentBlocks = append(currentBlocks, block)
		currentDividers = append(currentDividers, false)
	}
	flush()

	// 后处理：合并相邻同类型 section，但保留 divider 边界
	sections = sm.mergeAdjacentSameType(sections)

	// 后处理：提取标题、计算置信度
	for i := range sections {
		sections[i].Title = extractTitle(sections[i])
		sections[i].Confidence = sm.computeConfidence(sections[i])
	}

	return sections
}

func (sm *StateMachine) buildSection(secType SectionType, blocks []TextBlock, dividers []bool) Section {
	if len(blocks) == 0 {
		return Section{ID: "", Type: secType}
	}

	pageStart := blocks[0].PageNo
	pageEnd := blocks[0].PageNo
	var texts []string
	for _, b := range blocks {
		if b.PageNo < pageStart {
			pageStart = b.PageNo
		}
		if b.PageNo > pageEnd {
			pageEnd = b.PageNo
		}
		texts = append(texts, b.Text)
	}

	// 记录该 section 的第一个 block 是否由 divider 触发
	isDividerStart := false
	if len(dividers) > 0 {
		isDividerStart = dividers[0]
	}

	return Section{
		ID:             fmt.Sprintf("sec_%s_%04d", secType, 0),
		Type:           secType,
		Blocks:         append([]TextBlock(nil), blocks...),
		Text:           strings.Join(texts, "\n\n"),
		PageStart:      pageStart,
		PageEnd:        pageEnd,
		IsDividerStart: isDividerStart,
	}
}

func (sm *StateMachine) mergeAdjacentSameType(sections []Section) []Section {
	if len(sections) == 0 {
		return sections
	}

	var result []Section
	current := sections[0]

	for i := 1; i < len(sections); i++ {
		next := sections[i]
		// 只合并非 divider 边界的相邻同类型 section
		if next.Type == current.Type && !next.IsDividerStart {
			current.Blocks = append(current.Blocks, next.Blocks...)
			current.Text = current.Text + "\n\n" + next.Text
			if next.PageEnd > current.PageEnd {
				current.PageEnd = next.PageEnd
			}
			if next.PageStart < current.PageStart {
				current.PageStart = next.PageStart
			}
		} else {
			result = append(result, current)
			current = next
		}
	}
	result = append(result, current)

	// 重新分配 ID
	counters := make(map[SectionType]int)
	for i := range result {
		counters[result[i].Type]++
		result[i].ID = fmt.Sprintf("sec_%s_%03d", result[i].Type, counters[result[i].Type])
	}
	return result
}

func (sm *StateMachine) computeConfidence(sec Section) float64 {
	// 锚点直接命中的 section 置信度高
	if len(sec.Blocks) == 0 {
		return 0
	}
	firstBlock := sec.Blocks[0]
	match := BestAnchorForBlock(firstBlock, sm.anchors)
	if match != nil {
		base := 0.7 + float64(match.Anchor.Priority)/500.0 // 0.7 ~ 0.9
		if base > 0.92 {
			base = 0.92
		}
		return base
	}
	// 猜测的类型置信度低
	if sec.Type != SectionUnknown {
		return 0.45
	}
	return 0.2
}

func extractTitle(sec Section) string {
	if len(sec.Blocks) == 0 {
		return ""
	}
	// 取第一个 block 的第一行作为标题
	firstBlock := sec.Blocks[0]
	if len(firstBlock.Lines) == 0 {
		return ""
	}
	title := firstBlock.Lines[0].Text
	// 截断过长的标题
	runes := []rune(title)
	if len(runes) > 40 {
		return string(runes[:40]) + "..."
	}
	return title
}
