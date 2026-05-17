package parser

import "testing"

func TestDefaultAnchors(t *testing.T) {
	anchors := CompileAnchors(DefaultAnchors())
	if len(anchors) == 0 {
		t.Fatal("expected non-empty anchors")
	}

	tests := []struct {
		input    string
		expected SectionType
	}{
		{"给定资料一", SectionMaterial},
		{"材料二", SectionMaterial},
		{"第一题", SectionQuestion},
		{"第 2 题", SectionQuestion},
		{"请根据给定资料", SectionQuestion},
		{"作答要求：", SectionQuestion},
		{"参考答案", SectionAnswer},
		{"评分标准", SectionAnswer},
		{"参考解析", SectionAnalysis},
	}

	for _, tc := range tests {
		block := TextBlock{Lines: []TextLine{{Text: tc.input}}}
		match := BestAnchorForBlock(block, anchors)
		if match == nil {
			t.Errorf("input %q: expected match, got nil", tc.input)
			continue
		}
		if match.Anchor.Type != tc.expected {
			t.Errorf("input %q: expected %s, got %s", tc.input, tc.expected, match.Anchor.Type)
		}
	}
}

func TestGuessSectionTypeFromText(t *testing.T) {
	if got := GuessSectionTypeFromText("请根据资料概括问题"); got != SectionQuestion {
		t.Errorf("expected question, got %s", got)
	}
	if got := GuessSectionTypeFromText("评分细则如下"); got != SectionAnswer {
		t.Errorf("expected answer, got %s", got)
	}
	if got := GuessSectionTypeFromText("思路点拨"); got != SectionAnalysis {
		t.Errorf("expected analysis, got %s", got)
	}
}
