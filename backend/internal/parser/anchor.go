package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Anchor 表示一个文档结构锚点规则。
type Anchor struct {
	Name      string      `json:"name"`
	Pattern   string      `json:"pattern"`  // 正则表达式字符串
	Type      SectionType `json:"type"`
	Priority  int         `json:"priority"` // 优先级，数字越大越优先
	IsDivider bool        `json:"is_divider"` // 是否为分隔锚点：true 表示即使类型相同也创建新 section
	compiled  *regexp.Regexp
}

// AnchorMatch 表示一次锚点匹配结果。
type AnchorMatch struct {
	Anchor     Anchor   `json:"anchor"`
	BlockID    string   `json:"block_id"`
	LineText   string   `json:"line_text"`
	StartIndex int      `json:"start_index"` // 匹配在 block.Text 中的起始位置
}

// DefaultAnchors 返回申论文档的默认锚点规则集。
// IsDivider=true 的锚点用于内部切分（如多个题目、答案内的子题号）。
func DefaultAnchors() []Anchor {
	return []Anchor{
		// === 材料区（类型锚点，非分隔） ===
		{Name: "给定资料N", Pattern: `^(给定)?\s*资料\s*[一二三四五1-5][、\.．\s]*|^材料\s*[一二三四五1-5][、\.．\s]*`, Type: SectionMaterial, Priority: 100},
		{Name: "资料N", Pattern: `^资料\s*[一二三四五1-5]\b`, Type: SectionMaterial, Priority: 95},
		{Name: "材料N", Pattern: `^材料\s*[一二三四五1-5]\b`, Type: SectionMaterial, Priority: 95},
		{Name: "给定资料", Pattern: `^给定资料\b`, Type: SectionMaterial, Priority: 90},
		{Name: "阅读材料", Pattern: `^阅读材料\b`, Type: SectionMaterial, Priority: 85},

		// === 题目区 - 分隔锚点（题号，总是创建新 section） ===
		{Name: "第N题", Pattern: `^第\s*[一二三四五1-5]\s*题|^试题\s*[一二三四五1-5]|^问题\s*[一二三四五1-5]`, Type: SectionQuestion, Priority: 100, IsDivider: true},
		{Name: "阿拉伯题号", Pattern: `^[1-9][0-9]?[\.．、]\s*`, Type: SectionQuestion, Priority: 95, IsDivider: true},
		{Name: "中文题号", Pattern: `^[一二三四五六七八九十][、．\.]\s*`, Type: SectionQuestion, Priority: 95, IsDivider: true},
		{Name: "括号题号", Pattern: `^[（\(][1-9][0-9]?[）\)]\s*`, Type: SectionQuestion, Priority: 95, IsDivider: true},

		// === 题目区 - 类型锚点（非分隔） ===
		{Name: "作答要求", Pattern: `作答要求|答题要求|要求\s*[：:]`, Type: SectionQuestion, Priority: 85},
		{Name: "请根据", Pattern: `^请根据|^请结合|^请围绕|^请概括|^请分析|^请提出|^请指出`, Type: SectionQuestion, Priority: 80},
		{Name: "字数限制", Pattern: `不超过\s*\d+\s*字|字数限制|字数要求`, Type: SectionQuestion, Priority: 75},
		{Name: "根据材料", Pattern: `^根据(上述|给定|上述给定)?材料|^结合(上述|给定)?材料`, Type: SectionQuestion, Priority: 78},

		// === 答案区 - 分隔锚点（答案内的子题号） ===
		{Name: "答案子题号", Pattern: `^第\s*[一二三四五1-5]\s*题\s*[：:]`, Type: SectionAnswer, Priority: 101, IsDivider: true},
		{Name: "答案阿拉伯题号", Pattern: `^[1-9][0-9]?[\.．、]\s*(?:答案|解析|要点|参考)`, Type: SectionAnswer, Priority: 96, IsDivider: true},
		{Name: "答案括号题号", Pattern: `^[（\(][1-9][0-9]?[）\)]\s*(?:答案|解析|要点)`, Type: SectionAnswer, Priority: 96, IsDivider: true},

		// === 答案区 - 类型锚点（非分隔） ===
		{Name: "参考答案", Pattern: `参考答案|参考要点|答案要点`, Type: SectionAnswer, Priority: 100},
		{Name: "评分标准", Pattern: `评分标准|评分细则|赋分标准|评分说明|评分参考`, Type: SectionAnswer, Priority: 95},
		{Name: "答案", Pattern: `^答案[：:]|^答[：:]`, Type: SectionAnswer, Priority: 85},

		// === 解析区 ===
		{Name: "解析", Pattern: `参考解析|试题解析|答案解析|思路点拨|命题意图`, Type: SectionAnalysis, Priority: 90},
		{Name: "考查", Pattern: `考查能力|考查要点|知识拓展`, Type: SectionAnalysis, Priority: 80},
	}
}

// CompileAnchors 编译锚点规则的正则表达式。
func CompileAnchors(anchors []Anchor) []Anchor {
	result := make([]Anchor, 0, len(anchors))
	for _, a := range anchors {
		re, err := regexp.Compile(a.Pattern)
		if err != nil {
			continue // 跳过无效正则
		}
		a.compiled = re
		result = append(result, a)
	}
	return result
}

// MatchAnchors 对所有 block 执行锚点匹配。
func MatchAnchors(blocks []TextBlock, anchors []Anchor) []AnchorMatch {
	var matches []AnchorMatch
	for _, block := range blocks {
		if block.Meta.IsHeader || block.Meta.IsFooter || block.Meta.IsPageNumber {
			continue
		}
		// 优先匹配 block 的首行（锚点通常出现在段首）
		for _, line := range block.Lines {
			for _, anchor := range anchors {
				if anchor.compiled == nil {
					continue
				}
				if loc := anchor.compiled.FindStringIndex(line.Text); loc != nil {
					matches = append(matches, AnchorMatch{
						Anchor:     anchor,
						BlockID:    block.ID,
						LineText:   line.Text,
						StartIndex: loc[0],
					})
					// 一个 block 只匹配最高优先级的锚点（按类型去重）
					break
				}
			}
		}
	}
	return matches
}

// BestAnchorForBlock 返回对单个 block 的最佳锚点匹配（如果有）。
// 策略：优先只匹配第一行（锚点通常出现在段首），避免正文内容中的关键词被误识别。
func BestAnchorForBlock(block TextBlock, anchors []Anchor) *AnchorMatch {
	if len(block.Lines) == 0 {
		return nil
	}

	// 优先只匹配第一行
	var best *AnchorMatch
	for _, anchor := range anchors {
		if anchor.compiled == nil {
			continue
		}
		if loc := anchor.compiled.FindStringIndex(block.Lines[0].Text); loc != nil {
			if best == nil || anchor.Priority > best.Anchor.Priority {
				best = &AnchorMatch{
					Anchor:     anchor,
					BlockID:    block.ID,
					LineText:   block.Lines[0].Text,
					StartIndex: loc[0],
				}
			}
		}
	}
	if best != nil {
		return best
	}

	// 第一行无匹配时，再检查其他行（兼容性回退）
	for i := 1; i < len(block.Lines); i++ {
		for _, anchor := range anchors {
			if anchor.compiled == nil {
				continue
			}
			if loc := anchor.compiled.FindStringIndex(block.Lines[i].Text); loc != nil {
				return &AnchorMatch{
					Anchor:     anchor,
					BlockID:    block.ID,
					LineText:   block.Lines[i].Text,
					StartIndex: loc[0],
				}
			}
		}
	}
	return nil
}

// SplitBlocksByDividers 按 block 内部的 divider 锚点切分 block。
// 策略：先判断 block 的主类型（由首行锚点决定），然后只使用该类型对应的 divider 进行切分。
// 这避免了答案区内的 "第一题" 被当作 question divider 切分。
func SplitBlocksByDividers(blocks []TextBlock, anchors []Anchor) []TextBlock {
	anchors = CompileAnchors(anchors)
	var result []TextBlock
	for _, block := range blocks {
		subBlocks := splitSingleBlockByDividers(block, anchors)
		result = append(result, subBlocks...)
	}
	return result
}

func splitSingleBlockByDividers(block TextBlock, anchors []Anchor) []TextBlock {
	if len(block.Lines) <= 1 {
		return []TextBlock{block}
	}

	// 判断 block 的主类型
	mainMatch := BestAnchorForBlock(block, anchors)
	var mainType SectionType
	if mainMatch != nil {
		mainType = mainMatch.Anchor.Type
	}

	var subBlocks []TextBlock
	var current []TextLine
	flush := func() {
		if len(current) == 0 {
			return
		}
		subBlocks = append(subBlocks, TextBlock{
			ID:      fmt.Sprintf("%s_s%d", block.ID, len(subBlocks)),
			PageNo:  current[0].PageNo,
			Lines:   append([]TextLine(nil), current...),
			Text:    joinLines(current),
			Meta:    block.Meta,
		})
		current = nil
	}

	for _, line := range block.Lines {
		isDividerLine := false
		for _, anchor := range anchors {
			if !anchor.IsDivider || anchor.compiled == nil {
				continue
			}
			// 只使用与 block 主类型相同的 divider 进行切分
			if mainType != "" && anchor.Type != mainType {
				continue
			}
			if anchor.compiled.MatchString(line.Text) {
				isDividerLine = true
				break
			}
		}
		if isDividerLine && len(current) > 0 {
			flush()
		}
		current = append(current, line)
	}
	flush()

	if len(subBlocks) <= 1 {
		return []TextBlock{block}
	}
	return subBlocks
}

// GuessSectionTypeFromText 根据文本内容猜测 section 类型（无锚点时备用）。
func GuessSectionTypeFromText(text string) SectionType {
	text = strings.ToLower(text)
	if strings.Contains(text, "参考答案") || strings.Contains(text, "评分") || strings.Contains(text, "赋分") {
		return SectionAnswer
	}
	if strings.Contains(text, "解析") || strings.Contains(text, "思路") {
		return SectionAnalysis
	}
	if strings.Contains(text, "请根据") || strings.Contains(text, "请结合") || strings.Contains(text, "不超过") {
		return SectionQuestion
	}
	if strings.Contains(text, "资料") || strings.Contains(text, "材料") {
		return SectionMaterial
	}
	return SectionUnknown
}
