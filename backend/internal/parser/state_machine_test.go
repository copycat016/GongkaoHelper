package parser

import "testing"

func TestStateMachineBasic(t *testing.T) {
	anchors := CompileAnchors(DefaultAnchors())
	sm := NewStateMachine(anchors)

	blocks := []TextBlock{
		{ID: "b1", PageNo: 1, Lines: []TextLine{{Text: "给定资料一"}, {Text: "基层治理材料内容。"}}},
		{ID: "b2", PageNo: 1, Lines: []TextLine{{Text: "给定资料二"}, {Text: "公共服务材料内容。"}}},
		{ID: "b3", PageNo: 1, Lines: []TextLine{{Text: "第一题"}, {Text: "请根据给定资料概括问题。不超过300字。"}}},
		{ID: "b4", PageNo: 1, Lines: []TextLine{{Text: "第二题"}, {Text: "请提出对策建议。不超过500字。"}}},
		{ID: "b5", PageNo: 1, Lines: []TextLine{{Text: "参考答案"}, {Text: "第一题要点：..."}}},
		{ID: "b6", PageNo: 1, Lines: []TextLine{{Text: "评分标准"}, {Text: "每点4分..."}}},
	}

	for i := range blocks {
		blocks[i].Text = joinLines(blocks[i].Lines)
	}

	sections := sm.Run(blocks)
	// 题号锚点是 divider，相邻同类型的题目不会被合并
	// 材料1+材料2 -> material（非 divider，合并）
	// 题目1、题目2 -> 2 个独立 question section
	// 答案+评分 -> answer（非 divider，合并）
	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d: %+v", len(sections), sectionTypes(sections))
	}

	expected := []SectionType{SectionMaterial, SectionQuestion, SectionQuestion, SectionAnswer}
	for i, exp := range expected {
		if sections[i].Type != exp {
			t.Errorf("section %d: expected %s, got %s", i, exp, sections[i].Type)
		}
	}
}

func sectionTypes(sections []Section) []SectionType {
	var types []SectionType
	for _, s := range sections {
		types = append(types, s.Type)
	}
	return types
}
