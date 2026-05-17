package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LLMBoundaryResult 是 LLM 边界判断的输出格式。
type LLMBoundaryResult struct {
	SectionType  string  `json:"section_type"`
	StartBlockID string  `json:"start_block_id"`
	EndBlockID   string  `json:"end_block_id"`
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
}

// LLMRefiner 使用 LLM 精修状态机切分结果。
// 只在置信度低于阈值的边界处调用 LLM，不重新生成全文。
type LLMRefiner struct {
	client     *http.Client
	provider   LLMProviderConfig
	threshold  float64 // 低于此值的边界需要精修
	contextBlocks int  // 边界两侧各取多少 block
}

// LLMProviderConfig 用于调用 LLM 的最小配置。
type LLMProviderConfig struct {
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	TimeoutSec int   `json:"timeout_sec"`
}

// NewLLMRefiner 创建精修器。
func NewLLMRefiner(provider LLMProviderConfig) *LLMRefiner {
	timeout := time.Duration(provider.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &LLMRefiner{
		client:        &http.Client{Timeout: timeout},
		provider:      provider,
		threshold:     0.65,
		contextBlocks: 4,
	}
}

// Refine 对低置信度 section 的边界进行精修。
func (r *LLMRefiner) Refine(sections []Section) ([]Section, error) {
	if !r.isAvailable() {
		// LLM 不可用时，直接返回原结果
		return sections, nil
	}

	var refined []Section
	for i, sec := range sections {
		if sec.Confidence >= r.threshold || sec.Type == SectionUnknown {
			refined = append(refined, sec)
			continue
		}

		// 获取边界上下文
		contextBlocks := r.collectContextBlocks(sections, i)
		if len(contextBlocks) == 0 {
			refined = append(refined, sec)
			continue
		}

		// 调用 LLM 判断
		result, err := r.callLLMBoundary(contextBlocks, sec)
		if err != nil {
			// LLM 调用失败时，保持原 section 但降低置信度标记
			sec.Reason = fmt.Sprintf("LLM refinement failed: %v", err)
			refined = append(refined, sec)
			continue
		}

		// 根据 LLM 结果更新 section
		newSec := r.applyLLMResult(sec, result, contextBlocks)
		refined = append(refined, newSec)
	}

	return refined, nil
}

func (r *LLMRefiner) isAvailable() bool {
	return strings.TrimSpace(r.provider.BaseURL) != "" &&
		strings.TrimSpace(r.provider.APIKey) != "" &&
		strings.TrimSpace(r.provider.Model) != ""
}

// collectContextBlocks 收集边界附近需要 LLM 判断的 block。
func (r *LLMRefiner) collectContextBlocks(sections []Section, index int) []TextBlock {
	var blocks []TextBlock

	// 前一个 section 的尾部 block
	if index > 0 {
		prev := sections[index-1]
		start := max(0, len(prev.Blocks)-r.contextBlocks)
		for _, b := range prev.Blocks[start:] {
			blocks = append(blocks, b)
		}
	}

	// 当前 section 的全部 block
	blocks = append(blocks, sections[index].Blocks...)

	// 后一个 section 的头部 block
	if index < len(sections)-1 {
		next := sections[index+1]
		end := min(r.contextBlocks, len(next.Blocks))
		blocks = append(blocks, next.Blocks[:end]...)
	}

	return blocks
}

func (r *LLMRefiner) callLLMBoundary(blocks []TextBlock, target Section) (*LLMBoundaryResult, error) {
	prompt := r.buildBoundaryPrompt(blocks, target)

	payload := map[string]any{
		"model": r.provider.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个申论文档结构分析助手。你只负责判断文档边界，不要改写或总结文本内容。必须严格输出 JSON 格式。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.1,
		"max_tokens": 512,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", r.provider.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.provider.APIKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm api returned %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, errors.New("no choices from llm")
	}

	content := result.Choices[0].Message.Content
	// 尝试从 markdown code block 中提取 JSON
	content = extractJSONFromMarkdown(content)

	var boundary LLMBoundaryResult
	if err := json.Unmarshal([]byte(content), &boundary); err != nil {
		return nil, fmt.Errorf("failed to parse llm json: %w", err)
	}

	// 校验 section_type
	boundary.SectionType = normalizeSectionType(boundary.SectionType)
	return &boundary, nil
}

func (r *LLMRefiner) buildBoundaryPrompt(blocks []TextBlock, target Section) string {
	var sb strings.Builder
	sb.WriteString("以下是一段申论文档的连续文本片段，每个 block 用 [BLOCK:id] 标记。\n")
	sb.WriteString("当前状态机判断这是一个 '" + string(target.Type) + "' 区域，但置信度较低。\n")
	sb.WriteString("请判断：从哪个 block 开始真正进入了新的内容区域？\n")
	sb.WriteString("可选区域类型：material（给定资料/材料）、question（题目/问题）、answer（参考答案/评分标准）、analysis（解析/思路）。\n\n")

	for _, b := range blocks {
		sb.WriteString(fmt.Sprintf("[BLOCK:%s] page=%d\n%s\n\n", b.ID, b.PageNo, b.Text))
	}

	sb.WriteString("请输出 JSON，不要输出任何其他内容：\n")
	sb.WriteString(`{"section_type": "material|question|answer|analysis", "start_block_id": "b0001", "end_block_id": "b0005", "confidence": 0.85, "reason": "从 b0003 开始出现'给定资料一'，标志着进入材料区"}`)
	return sb.String()
}

func (r *LLMRefiner) applyLLMResult(sec Section, result *LLMBoundaryResult, context []TextBlock) Section {
	sec.Reason = fmt.Sprintf("LLM refined: %s (confidence=%.2f)", result.Reason, result.Confidence)

	// 如果 LLM 判断的类型与当前不同，更新类型
	newType := SectionType(result.SectionType)
	if newType != "" && newType != sec.Type {
		sec.Type = newType
	}

	// 根据 start_block_id 和 end_block_id 调整 block 范围
	if result.StartBlockID != "" {
		var filtered []TextBlock
		started := false
		for _, b := range sec.Blocks {
			if b.ID == result.StartBlockID {
				started = true
			}
			if started {
				filtered = append(filtered, b)
			}
			if b.ID == result.EndBlockID {
				break
			}
		}
		if len(filtered) > 0 {
			sec.Blocks = filtered
			var texts []string
			for _, b := range filtered {
				texts = append(texts, b.Text)
			}
			sec.Text = strings.Join(texts, "\n\n")
		}
	}

	// 更新置信度
	if result.Confidence > 0 && result.Confidence <= 1.0 {
		sec.Confidence = result.Confidence
	} else {
		sec.Confidence = 0.7
	}

	return sec
}

func extractJSONFromMarkdown(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				continue
			}
			result = append(result, line)
		}
		return strings.Join(result, "\n")
	}
	return content
}

func normalizeSectionType(t string) string {
	switch strings.ToLower(t) {
	case "material", "材料":
		return "material"
	case "question", "题目", "问题", "试题":
		return "question"
	case "answer", "答案", "参考答案", "评分标准":
		return "answer"
	case "analysis", "解析", "思路", "思路点拨":
		return "analysis"
	default:
		return "unknown"
	}
}
