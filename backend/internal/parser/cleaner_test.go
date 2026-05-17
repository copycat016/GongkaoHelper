package parser

import (
	"testing"
)

func TestCleanerRemovePageNumbers(t *testing.T) {
	c := NewCleaner()
	blocks := []TextBlock{
		{ID: "b1", PageNo: 1, Lines: []TextLine{{Text: "第 1 页"}}},
		{ID: "b2", PageNo: 1, Lines: []TextLine{{Text: "材料一"}, {Text: "这是材料内容。"}}},
		{ID: "b3", PageNo: 1, Lines: []TextLine{{Text: "- 2 -"}}},
	}
	result := c.Clean(blocks)
	if len(result) != 1 {
		t.Fatalf("expected 1 block after cleaning page numbers, got %d", len(result))
	}
	if result[0].ID != "b2" {
		t.Errorf("expected remaining block b2, got %s", result[0].ID)
	}
}

func TestCleanerDeduplicateLines(t *testing.T) {
	c := NewCleaner()
	blocks := []TextBlock{
		{ID: "b1", PageNo: 1, Lines: []TextLine{
			{Text: "重复行"},
			{Text: "重复行"},
			{Text: "重复行"},
			{Text: "不重复"},
		}},
	}
	result := c.Clean(blocks)
	lines := result[0].Lines
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines after dedup, got %d", len(lines))
	}
	if lines[0].Text != "重复行" || lines[1].Text != "不重复" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestCleanerMergeBrokenLines(t *testing.T) {
	c := NewCleaner()
	blocks := []TextBlock{
		{ID: "b1", PageNo: 1, Lines: []TextLine{{Text: "这是一个因为排版原因"}}},
		{ID: "b2", PageNo: 1, Lines: []TextLine{{Text: "被断开的句子。"}}},
		{ID: "b3", PageNo: 1, Lines: []TextLine{{Text: "第二段开始了。"}}},
	}
	result := c.Clean(blocks)
	// b1 和 b2 应该合并，b3 独立
	if len(result) != 2 {
		t.Fatalf("expected 2 blocks after merge, got %d", len(result))
	}
	if !contains(result[0].Text, "被断开的句子") {
		t.Errorf("expected merged text containing '被断开的句子', got: %s", result[0].Text)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSub(s, substr))
}

func containsSub(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
