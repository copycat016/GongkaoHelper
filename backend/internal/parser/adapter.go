package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OCRAdapter 定义 OCR/PDF 输出到 parser 输入的适配器接口。
// 方便后续接入不同 OCR 服务（百度、PaddleOCR、VL 模型等）。
type OCRAdapter interface {
	// Adapt 将原始 OCR/PDF 输出转换为 parser 可处理的 TextLine 列表。
	Adapt(raw any) ([]TextLine, error)
	// Name 返回适配器名称。
	Name() string
}

// RawLine 是通用的原始行输入，各适配器统一输出此格式。
type RawLine struct {
	ID      string `json:"id,omitempty"`
	BlockID string `json:"block_id,omitempty"`
	Text    string `json:"text"`
	PageNo  int    `json:"page_no"`
	// 可选的位置信息（如 OCR 返回了坐标）
	Left   float64 `json:"left,omitempty"`
	Top    float64 `json:"top,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// RawDocument 是 parser 的通用 OCR/PDF 输入结构。
// 后续无论接百度 OCR、PaddleOCR 还是视觉模型，都先转换成 page/block/line。
type RawDocument struct {
	Pages []RawPage `json:"pages"`
}

type RawPage struct {
	PageNo int        `json:"page_no"`
	Blocks []RawBlock `json:"blocks"`
}

type RawBlock struct {
	ID    string    `json:"id,omitempty"`
	Lines []RawLine `json:"lines"`
}

// PlainTextAdapter 适配纯文本输入（如 PDF 提取的文本）。
type PlainTextAdapter struct{}

func NewPlainTextAdapter() *PlainTextAdapter { return &PlainTextAdapter{} }

func (a *PlainTextAdapter) Name() string { return "plain_text" }

func (a *PlainTextAdapter) Adapt(raw any) ([]TextLine, error) {
	text, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("plain text adapter expects string input, got %T", raw)
	}

	// 支持分页标记：--- page N --- 或 \f
	pages := splitTextToPages(text)
	var lines []TextLine
	lineCounter := 0

	for _, page := range pages {
		for _, lineText := range strings.Split(page.Text, "\n") {
			trimmed := strings.TrimSpace(lineText)
			// 保留空行作为 block 分割信号；BuildBlocks 会利用空行进行分段
			lines = append(lines, TextLine{
				ID:     fmt.Sprintf("l%05d", lineCounter),
				Text:   trimmed,
				PageNo: page.PageNo,
			})
			lineCounter++
		}
	}
	return lines, nil
}

// StructuredDocumentAdapter 适配已经保留 page/block/line 的 OCR/PDF 输出。
type StructuredDocumentAdapter struct{}

func NewStructuredDocumentAdapter() *StructuredDocumentAdapter { return &StructuredDocumentAdapter{} }

func (a *StructuredDocumentAdapter) Name() string { return "structured_document" }

func (a *StructuredDocumentAdapter) Adapt(raw any) ([]TextLine, error) {
	doc, err := normalizeRawDocument(raw)
	if err != nil {
		return nil, err
	}
	var lines []TextLine
	lineCounter := 0
	for pageIndex, page := range doc.Pages {
		pageNo := page.PageNo
		if pageNo <= 0 {
			pageNo = pageIndex + 1
		}
		for blockIndex, block := range page.Blocks {
			blockID := block.ID
			if blockID == "" {
				blockID = fmt.Sprintf("p%03d_b%04d", pageNo, blockIndex)
			}
			for _, line := range block.Lines {
				lineID := line.ID
				if lineID == "" {
					lineID = fmt.Sprintf("l%05d", lineCounter)
				}
				linePageNo := line.PageNo
				if linePageNo <= 0 {
					linePageNo = pageNo
				}
				lineBlockID := line.BlockID
				if lineBlockID == "" {
					lineBlockID = blockID
				}
				lines = append(lines, TextLine{
					ID:      lineID,
					BlockID: lineBlockID,
					Text:    strings.TrimSpace(line.Text),
					PageNo:  linePageNo,
				})
				lineCounter++
			}
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("structured document has no lines")
	}
	return lines, nil
}

func normalizeRawDocument(raw any) (RawDocument, error) {
	switch value := raw.(type) {
	case RawDocument:
		return value, nil
	case *RawDocument:
		if value == nil {
			return RawDocument{}, fmt.Errorf("structured document is nil")
		}
		return *value, nil
	case []byte:
		var doc RawDocument
		if err := json.Unmarshal(value, &doc); err != nil {
			return RawDocument{}, fmt.Errorf("decode structured document failed: %w", err)
		}
		return doc, nil
	case string:
		var doc RawDocument
		if err := json.Unmarshal([]byte(value), &doc); err != nil {
			return RawDocument{}, fmt.Errorf("decode structured document failed: %w", err)
		}
		return doc, nil
	default:
		return RawDocument{}, fmt.Errorf("structured document adapter expects RawDocument or JSON, got %T", raw)
	}
}

type pageText struct {
	PageNo int
	Text   string
}

func splitTextToPages(text string) []pageText {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// 先尝试按分页符分割
	if strings.Contains(text, "\f") {
		parts := strings.Split(text, "\f")
		var pages []pageText
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" && i == len(parts)-1 {
				continue
			}
			pages = append(pages, pageText{PageNo: i + 1, Text: p})
		}
		if len(pages) > 0 {
			return pages
		}
	}

	// 尝试按 "--- page N ---" 标记分割
	var pages []pageText
	var currentLines []string
	currentPage := 1

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if isPageMarker(trimmed) {
			if len(currentLines) > 0 {
				pages = append(pages, pageText{PageNo: currentPage, Text: strings.Join(currentLines, "\n")})
				currentLines = nil
			}
			currentPage = extractPageNo(trimmed)
			continue
		}
		currentLines = append(currentLines, line)
	}
	if len(currentLines) > 0 {
		pages = append(pages, pageText{PageNo: currentPage, Text: strings.Join(currentLines, "\n")})
	}

	if len(pages) == 0 && strings.TrimSpace(text) != "" {
		return []pageText{{PageNo: 1, Text: text}}
	}
	return pages
}

func isPageMarker(line string) bool {
	line = strings.ToLower(line)
	return strings.HasPrefix(line, "--- page ") && strings.HasSuffix(line, " ---")
}

func extractPageNo(line string) int {
	var page int
	fmt.Sscanf(line, "--- page %d ---", &page)
	if page <= 0 {
		fmt.Sscanf(line, "--- Page %d ---", &page)
	}
	if page <= 0 {
		page = 1
	}
	return page
}

// BaiduOCRAdapter 适配百度 OCR 输出（words_result 列表）。
type BaiduOCRAdapter struct{}

func NewBaiduOCRAdapter() *BaiduOCRAdapter { return &BaiduOCRAdapter{} }

func (a *BaiduOCRAdapter) Name() string { return "baidu_ocr" }

func (a *BaiduOCRAdapter) Adapt(raw any) ([]TextLine, error) {
	// 百度 OCR 通用文字识别返回 words_result 数组，每个元素有 words 字段。
	// 多页/复杂版建议在调用层先转换成 RawDocument，再交给 structured_document。
	type baiduResult struct {
		WordsResult []struct {
			Words    string `json:"words"`
			Location struct {
				Left   float64 `json:"left"`
				Top    float64 `json:"top"`
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
			} `json:"location"`
		} `json:"words_result"`
	}

	var body []byte
	switch value := raw.(type) {
	case []byte:
		body = value
	case string:
		body = []byte(value)
	default:
		return nil, fmt.Errorf("baidu ocr adapter expects []byte input, got %T", raw)
	}

	var result baiduResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode baidu ocr result failed: %w", err)
	}

	lines := make([]TextLine, 0, len(result.WordsResult))
	for index, item := range result.WordsResult {
		text := strings.TrimSpace(item.Words)
		if text == "" {
			continue
		}
		blockID := fmt.Sprintf("p001_b%04d", index)
		lines = append(lines, TextLine{
			ID:      fmt.Sprintf("l%05d", index),
			BlockID: blockID,
			Text:    text,
			PageNo:  1,
		})
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("baidu ocr result has no words_result lines")
	}
	return lines, nil
}

// AdapterRegistry 管理所有适配器。
type AdapterRegistry struct {
	adapters map[string]OCRAdapter
}

func NewAdapterRegistry() *AdapterRegistry {
	r := &AdapterRegistry{adapters: make(map[string]OCRAdapter)}
	r.Register(NewPlainTextAdapter())
	r.Register(NewStructuredDocumentAdapter())
	r.Register(NewBaiduOCRAdapter())
	return r
}

func (r *AdapterRegistry) Register(adapter OCRAdapter) {
	r.adapters[adapter.Name()] = adapter
}

func (r *AdapterRegistry) Get(name string) (OCRAdapter, bool) {
	a, ok := r.adapters[name]
	return a, ok
}
