package services

import (
	"fmt"
	"sort"
	"strings"

	"gkweb/backend/internal/parser"
)

type DocumentParseInput struct {
	RawText    string
	PDFPath    string
	OCRAdapter string
	OCRRaw     any
}

type DocumentParseResult struct {
	Source  string         `json:"source"`
	Text    string         `json:"text"`
	Pages   []PDFTextPage  `json:"pages,omitempty"`
	Quality PDFTextQuality `json:"quality"`
	Lines   int            `json:"line_count"`
}

// ParseDocumentSource is the unified entry for text-like document input.
// PDF text layers, OCR JSON, structured OCR output and pasted raw text are
// normalized into the same plain text shape consumed by the essay parser.
func ParseDocumentSource(input DocumentParseInput) (*DocumentParseResult, error) {
	if text := strings.TrimSpace(input.RawText); text != "" {
		return parseRawTextDocument(text), nil
	}

	if path := strings.TrimSpace(input.PDFPath); path != "" {
		pages, err := ExtractPDFTextPages(path)
		if err != nil {
			return nil, err
		}
		text := pagesToEssayText(pages)
		return &DocumentParseResult{
			Source:  "pdf_text",
			Text:    text,
			Pages:   pages,
			Quality: PDFTextQuality{OK: true, Reason: "pdf text layer looks usable"},
			Lines:   countTextLines(text),
		}, nil
	}

	if input.OCRRaw != nil {
		adapterName := strings.TrimSpace(input.OCRAdapter)
		if adapterName == "" {
			adapterName = "baidu_ocr"
		}
		registry := parser.NewAdapterRegistry()
		adapter, ok := registry.Get(adapterName)
		if !ok {
			return nil, fmt.Errorf("unknown ocr/pdf adapter: %s", adapterName)
		}
		lines, err := adapter.Adapt(input.OCRRaw)
		if err != nil {
			return nil, err
		}
		text := textLinesToEssayText(lines)
		return &DocumentParseResult{
			Source: adapterName,
			Text:   text,
			Quality: PDFTextQuality{
				OK:     strings.TrimSpace(text) != "",
				Reason: "ocr output normalized",
			},
			Lines: len(lines),
		}, nil
	}

	return nil, fmt.Errorf("raw_text, pdf file or ocr output is required")
}

func parseRawTextDocument(text string) *DocumentParseResult {
	return &DocumentParseResult{
		Source:  "raw_text",
		Text:    text,
		Quality: PDFTextQuality{OK: true, Reason: "raw text provided"},
		Lines:   countTextLines(text),
	}
}

func textLinesToEssayText(lines []parser.TextLine) string {
	type pageLines struct {
		pageNo int
		lines  []string
	}
	pageMap := make(map[int][]string)
	for _, line := range lines {
		text := strings.TrimSpace(line.Text)
		if text == "" {
			continue
		}
		pageNo := line.PageNo
		if pageNo <= 0 {
			pageNo = 1
		}
		pageMap[pageNo] = append(pageMap[pageNo], text)
	}

	pages := make([]pageLines, 0, len(pageMap))
	for pageNo, items := range pageMap {
		pages = append(pages, pageLines{pageNo: pageNo, lines: items})
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].pageNo < pages[j].pageNo
	})

	parts := make([]string, 0, len(pages))
	for _, page := range pages {
		parts = append(parts, fmt.Sprintf("--- page %d ---\n%s", page.pageNo, strings.Join(page.lines, "\n")))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func countTextLines(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
