package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

type PDFTextPage struct {
	PageNo int
	Text   string
}

type PDFTextQuality struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

func ExtractPDFTextPages(path string) ([]PDFTextPage, error) {
	pages, quality, err := ExtractPDFTextPagesForTest(path)
	if err != nil {
		return nil, err
	}
	if !quality.OK {
		return pages, fmt.Errorf("pdf text layer is not reliably decodable: %s; please use OCR", quality.Reason)
	}
	return pages, nil
}

func ExtractPDFTextPagesForTest(path string) ([]PDFTextPage, PDFTextQuality, error) {
	if pages, quality, err := extractPDFTextPagesWithPoppler(path); err == nil {
		return pages, quality, nil
	}
	return extractPDFTextPagesWithGo(path)
}

func extractPDFTextPagesWithGo(path string) ([]PDFTextPage, PDFTextQuality, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return nil, PDFTextQuality{}, err
	}
	defer file.Close()

	totalPages := reader.NumPage()
	if totalPages <= 0 {
		return nil, PDFTextQuality{}, errors.New("pdf has no pages")
	}

	fonts := make(map[string]*pdf.Font)
	pages := make([]PDFTextPage, 0, totalPages)
	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() || page.V.Key("Contents").Kind() == pdf.Null {
			pages = append(pages, PDFTextPage{PageNo: pageIndex})
			continue
		}

		for _, name := range page.Fonts() {
			if _, ok := fonts[name]; !ok {
				font := page.Font(name)
				fonts[name] = &font
			}
		}

		text, err := page.GetPlainText(fonts)
		if err != nil {
			return nil, PDFTextQuality{}, err
		}
		pages = append(pages, PDFTextPage{
			PageNo: pageIndex,
			Text:   normalizePDFText(text),
		})
	}

	if countExtractedText(pages) == 0 {
		return pages, PDFTextQuality{OK: false, Reason: "no extractable text found in pdf, it may be a scanned document"}, nil
	}
	return pages, assessPDFTextQuality(pages), nil
}

func extractPDFTextPagesWithPoppler(path string) ([]PDFTextPage, PDFTextQuality, error) {
	binary, err := exec.LookPath("pdftotext")
	if err != nil {
		return nil, PDFTextQuality{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "-layout", "-enc", "UTF-8", path, "-")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, PDFTextQuality{}, fmt.Errorf("pdftotext failed: %s", strings.TrimSpace(stderr.String()))
	}

	pages := textToPDFPages(string(output))
	if countExtractedText(pages) == 0 {
		return pages, PDFTextQuality{OK: false, Reason: "pdftotext found no extractable text"}, nil
	}
	return pages, assessPDFTextQuality(pages), nil
}

func textToPDFPages(text string) []PDFTextPage {
	text = sanitizePostgresText(text)
	parts := strings.Split(text, "\f")
	if len(parts) == 0 {
		parts = []string{text}
	}
	pages := make([]PDFTextPage, 0, len(parts))
	for index, part := range parts {
		part = normalizePDFText(part)
		if part == "" && index == len(parts)-1 {
			continue
		}
		pages = append(pages, PDFTextPage{
			PageNo: index + 1,
			Text:   part,
		})
	}
	if len(pages) == 0 {
		return []PDFTextPage{{PageNo: 1}}
	}
	return pages
}

func pagesToEssayText(pages []PDFTextPage) string {
	parts := make([]string, 0, len(pages))
	for _, page := range pages {
		parts = append(parts, "--- page "+itoa(page.PageNo)+" ---\n"+page.Text)
	}
	return strings.Join(parts, "\n\n")
}

func normalizePDFText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = sanitizePostgresText(text)
	lines := strings.Split(text, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			normalized = append(normalized, "")
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.TrimSpace(strings.Join(normalized, "\n"))
}

func sanitizePostgresText(text string) string {
	text = strings.ReplaceAll(text, "\x00", "")
	if utf8.ValidString(text) {
		return text
	}
	return strings.ToValidUTF8(text, "")
}

func countExtractedText(pages []PDFTextPage) int {
	count := 0
	for _, page := range pages {
		count += len([]rune(strings.TrimSpace(page.Text)))
	}
	return count
}

func assessPDFTextQuality(pages []PDFTextPage) PDFTextQuality {
	var total, cjk, asciiLetters, digits, spaces, punctuation, bad int
	for _, page := range pages {
		for _, r := range page.Text {
			total++
			switch {
			case isCJK(r):
				cjk++
			case r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
				if unicode.IsDigit(r) {
					digits++
				} else {
					asciiLetters++
				}
			case unicode.IsSpace(r):
				spaces++
			case unicode.IsPunct(r) || strings.ContainsRune("，。！？；：（）《》、—“”‘’·…【】", r):
				punctuation++
			case r == utf8.RuneError || unicode.IsControl(r):
				bad++
			default:
				if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
					bad++
				}
			}
		}
	}
	if total < 40 {
		return PDFTextQuality{OK: false, Reason: "too little text was extracted"}
	}
	readable := cjk + asciiLetters + digits + spaces + punctuation
	cjkRatio := float64(cjk) / float64(total)
	badRatio := float64(bad) / float64(total)
	readableRatio := float64(readable) / float64(total)
	if badRatio > 0.18 {
		return PDFTextQuality{OK: false, Reason: fmt.Sprintf("too many undecodable characters (%.0f%%)", badRatio*100)}
	}
	if cjkRatio < 0.08 && readableRatio < 0.78 {
		return PDFTextQuality{OK: false, Reason: "Chinese text ratio is too low for an essay paper"}
	}
	return PDFTextQuality{OK: true, Reason: "text layer looks usable"}
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) ||
		(r >= 0x2B740 && r <= 0x2B81F) ||
		(r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0xF900 && r <= 0xFAFF)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
