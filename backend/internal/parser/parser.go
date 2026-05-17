package parser

import (
	"fmt"
)

// Config 是 Parser 的配置选项。
type Config struct {
	Anchors     []Anchor          // 自定义锚点规则（nil 则使用默认）
	Cleaner     *Cleaner          // 自定义清洗器（nil 则使用默认）
	UseLLM      bool              // 是否启用 LLM 边界精修
	LLMProvider LLMProviderConfig // LLM 配置
}

// Parser 是 PDF/OCR 文本解析的主入口。
type Parser struct {
	config Config
}

// New 创建 Parser 实例。
func New(cfg Config) *Parser {
	if cfg.Cleaner == nil {
		cfg.Cleaner = NewCleaner()
	}
	if cfg.Anchors == nil {
		cfg.Anchors = DefaultAnchors()
	}
	return &Parser{config: cfg}
}

// NewDefault 创建使用默认配置的 Parser。
func NewDefault() *Parser {
	return New(Config{
		Anchors: DefaultAnchors(),
		Cleaner: NewCleaner(),
		UseLLM:  false,
	})
}

// ParseString 从纯文本字符串解析（如 PDF 提取的文本）。
func (p *Parser) ParseString(documentID string, text string) (StructuredResult, error) {
	adapter := NewPlainTextAdapter()
	lines, err := adapter.Adapt(text)
	if err != nil {
		return StructuredResult{}, fmt.Errorf("adapt failed: %w", err)
	}
	return p.ParseLines(documentID, lines)
}

func (p *Parser) PrepareString(documentID string, text string) (CleanedDocument, error) {
	adapter := NewPlainTextAdapter()
	lines, err := adapter.Adapt(text)
	if err != nil {
		return CleanedDocument{}, fmt.Errorf("adapt failed: %w", err)
	}
	return p.PrepareLines(documentID, lines)
}

// PrepareLines 只做原始结构构建和清洗，不做语义切分。
func (p *Parser) PrepareLines(documentID string, lines []TextLine) (CleanedDocument, error) {
	if len(lines) == 0 {
		return CleanedDocument{}, fmt.Errorf("no text lines to parse")
	}

	blocks := BuildBlocks(lines)
	if len(blocks) == 0 {
		return CleanedDocument{}, fmt.Errorf("no blocks after building")
	}

	blocks = p.config.Cleaner.Clean(blocks)
	if len(blocks) == 0 {
		return CleanedDocument{}, fmt.Errorf("no blocks after cleaning")
	}

	return CleanedDocument{DocumentID: documentID, Blocks: blocks}, nil
}

// ParseLines 从已经构建好的 TextLine 列表解析。
func (p *Parser) ParseLines(documentID string, lines []TextLine) (StructuredResult, error) {
	cleaned, err := p.PrepareLines(documentID, lines)
	if err != nil {
		return StructuredResult{}, err
	}
	return p.ParsePrepared(cleaned)
}

// ParsePrepared 使用规则锚点和状态机切分，作为无 LLM 时的兜底路径。
func (p *Parser) ParsePrepared(cleaned CleanedDocument) (StructuredResult, error) {
	blocks := SplitBlocksByDividers(cleaned.Blocks, p.config.Anchors)

	// 3. 状态机切分 section
	sm := NewStateMachine(p.config.Anchors)
	sections := sm.Run(blocks)
	if len(sections) == 0 {
		return StructuredResult{}, fmt.Errorf("no sections after state machine")
	}

	// 4. LLM 精修（可选）
	if p.config.UseLLM {
		refiner := NewLLMRefiner(p.config.LLMProvider)
		sections, _ = refiner.Refine(sections)
	}

	// 5. 构建结构化结果
	result := BuildResult(cleaned.DocumentID, sections)
	return result, nil
}

// ParsePreparedWithBoundaryPlan 使用快速模型输出的边界计划切分。
func (p *Parser) ParsePreparedWithBoundaryPlan(cleaned CleanedDocument, plan BoundaryPlan) (StructuredResult, error) {
	sections, err := ApplyBoundaryPlan(cleaned.Blocks, plan)
	if err != nil {
		return StructuredResult{}, err
	}
	result := BuildResult(cleaned.DocumentID, sections)
	return result, nil
}

// ParseWithAdapter 使用指定适配器解析原始数据。
func (p *Parser) ParseWithAdapter(documentID string, adapterName string, raw any) (StructuredResult, error) {
	registry := NewAdapterRegistry()
	adapter, ok := registry.Get(adapterName)
	if !ok {
		return StructuredResult{}, fmt.Errorf("unknown adapter: %s", adapterName)
	}
	lines, err := adapter.Adapt(raw)
	if err != nil {
		return StructuredResult{}, fmt.Errorf("adapter %s failed: %w", adapterName, err)
	}
	return p.ParseLines(documentID, lines)
}

// Stats 返回解析过程的统计信息，便于调试。
type Stats struct {
	LineCount    int `json:"line_count"`
	BlockCount   int `json:"block_count"`
	SectionCount int `json:"section_count"`
	CleanedCount int `json:"cleaned_count"` // 被清洗器移除的 block 数
}

// ParseWithStats 解析并返回统计信息。
func (p *Parser) ParseWithStats(documentID string, lines []TextLine) (StructuredResult, Stats, error) {
	stats := Stats{LineCount: len(lines)}

	blocks := BuildBlocks(lines)
	stats.BlockCount = len(blocks)

	cleaned := p.config.Cleaner.Clean(blocks)
	stats.CleanedCount = len(blocks) - len(cleaned)
	cleaned = SplitBlocksByDividers(cleaned, p.config.Anchors)

	sm := NewStateMachine(p.config.Anchors)
	sections := sm.Run(cleaned)
	stats.SectionCount = len(sections)

	if p.config.UseLLM {
		refiner := NewLLMRefiner(p.config.LLMProvider)
		sections, _ = refiner.Refine(sections)
	}

	result := BuildResult(documentID, sections)
	return result, stats, nil
}
