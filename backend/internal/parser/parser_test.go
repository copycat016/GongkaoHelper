package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseStringEndToEnd(t *testing.T) {
	text := `材料一
基层治理材料内容，围绕公共服务和群众诉求展开。

材料二
有群众反映办事流程较长，部门之间协同不足。

第一题
请根据给定资料，概括基层治理中存在的主要问题。要求：全面、准确、有条理，不超过300字。20分

第二题
请结合给定资料，提出提升公共服务效率的对策建议。要求：措施具体、可操作，不超过500字。30分

参考答案
第一题参考要点：信息收集不充分、部门协同不足。
第二题参考要点：建立统一平台、压实部门责任。

评分标准
概括题每个要点4分，逻辑条理4分。对策题每条有效对策5分。`

	p := NewDefault()
	result, err := p.ParseString("doc_test", text)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(result.Sections) == 0 {
		t.Fatal("expected non-empty sections")
	}

	// 验证 BuildResult
	if len(result.Materials) == 0 {
		t.Error("expected at least one material")
	}
	if len(result.Questions) == 0 {
		t.Error("expected at least one question")
	}
	if len(result.Answers) == 0 {
		t.Error("expected at least one answer")
	}

	// 验证题目关联了材料
	for _, q := range result.Questions {
		if len(q.MaterialIDs) == 0 {
			t.Errorf("question %s should have material refs", q.ID)
		}
		if len(q.AnswerIDs) == 0 {
			t.Errorf("question %s should have answer refs", q.ID)
		}
	}

	// 验证 ContextForReview
	q := result.Questions[0]
	question, materials, answers, found := result.ContextForReview(q.ID)
	if !found {
		t.Fatal("expected question to be found")
	}
	if question.ID != q.ID {
		t.Errorf("expected question id %s, got %s", q.ID, question.ID)
	}
	if len(materials) == 0 {
		t.Error("expected materials in review context")
	}
	if len(answers) == 0 {
		t.Error("expected answers in review context")
	}
}

func TestParseMultipleQuestions(t *testing.T) {
	text := `给定资料1
材料内容A。

给定资料2
材料内容B。

一、请概括问题。不超过200字。10分

二、请分析原因。不超过300字。15分

三、请提出对策。不超过400字。20分

参考答案
一、要点A。
二、要点B。
三、要点C。`

	p := NewDefault()
	result, err := p.ParseString("doc_multi", text)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// 相邻同类型 material 合并为 1 个；中文题号是 divider，每道题独立；答案合并为 1 个
	questionCount := 0
	materialCount := 0
	answerCount := 0
	for _, sec := range result.Sections {
		switch sec.Type {
		case SectionQuestion:
			questionCount++
		case SectionMaterial:
			materialCount++
		case SectionAnswer:
			answerCount++
		}
	}

	if materialCount != 1 {
		t.Errorf("expected 1 merged material, got %d", materialCount)
	}
	if questionCount != 3 {
		t.Errorf("expected 3 questions, got %d", questionCount)
	}
	if answerCount != 1 {
		t.Errorf("expected 1 answer, got %d", answerCount)
	}

	// BuildResult 也应该生成 3 个 question
	if len(result.Questions) != 3 {
		t.Errorf("expected 3 questions in result, got %d", len(result.Questions))
	}
}

func TestParseArabicNumberQuestions(t *testing.T) {
	text := `1. 请概括问题。不超过200字。

2. 请分析原因。不超过300字。

3. 请提出对策。不超过400字。

答案：
1. 要点A。
2. 要点B。
3. 要点C。`

	p := NewDefault()
	result, err := p.ParseString("doc_arabic", text)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	questionCount := 0
	answerCount := 0
	for _, sec := range result.Sections {
		switch sec.Type {
		case SectionQuestion:
			questionCount++
		case SectionAnswer:
			answerCount++
		}
	}

	if questionCount != 3 {
		t.Errorf("expected 3 questions, got %d", questionCount)
	}
	if answerCount != 4 {
		// "答案：" + 3 个子题，共 4 个 answer section（divider 边界保留）
		t.Errorf("expected 4 answers, got %d", answerCount)
	}
}

func TestParseAnswerWithSubQuestions(t *testing.T) {
	text := `第一题
请概括问题。

第二题
请分析原因。

参考答案
第一题：信息收集不充分。
第二题：责任边界不清。

评分标准
第一题每点5分。
第二题每点10分。`

	p := NewDefault()
	result, err := p.ParseString("doc_answer_sub", text)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// 2 题 + 2 答案（参考答案里第一题和第二题是 divider）+ 1 评分（合并为一个）
	questionCount := 0
	answerCount := 0
	for _, sec := range result.Sections {
		fmt.Printf("  [%s] %s\n", sec.Type, sec.Title)
		switch sec.Type {
		case SectionQuestion:
			questionCount++
		case SectionAnswer:
			answerCount++
		}
	}

	if questionCount != 2 {
		t.Errorf("expected 2 questions, got %d", questionCount)
	}
	// 参考答案 + 评分标准 = 至少 2 个 answer section
	// 参考答案里的"第一题"和"第二题"由于优先匹配首行"参考答案"，不会额外切分
	if answerCount < 2 {
		t.Errorf("expected at least 2 answers, got %d", answerCount)
	}
}

func TestParseStringWithPageMarkers(t *testing.T) {
	text := "--- page 1 ---\n材料一\n内容A。\n\n--- page 2 ---\n第一题\n请概括问题。\n\n--- page 3 ---\n参考答案\n要点A。"
	p := NewDefault()
	result, err := p.ParseString("doc_paged", text)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	for _, sec := range result.Sections {
		if sec.PageStart < 1 {
			t.Errorf("section %s page_start should >= 1, got %d", sec.ID, sec.PageStart)
		}
	}
}

func TestParseStringEmptyInput(t *testing.T) {
	p := NewDefault()
	_, err := p.ParseString("doc_empty", "")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestAdapterPlainText(t *testing.T) {
	adapter := NewPlainTextAdapter()
	lines, err := adapter.Adapt("第一行\n第二行\n\n第三行")
	if err != nil {
		t.Fatalf("adapt failed: %v", err)
	}
	// 空行被保留作为 block 分割信号，所以是 4 行（包含一个空行）
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	if lines[0].Text != "第一行" {
		t.Errorf("unexpected first line: %s", lines[0].Text)
	}
}

func TestAdapterWithPageBreaks(t *testing.T) {
	adapter := NewPlainTextAdapter()
	input := "页1行1\n页1行2\n\f\n页2行1\n页2行2"
	lines, err := adapter.Adapt(input)
	if err != nil {
		t.Fatalf("adapt failed: %v", err)
	}
	// 第二页的行 page_no 应该为 2
	var page2Count int
	for _, l := range lines {
		if l.PageNo == 2 {
			page2Count++
		}
	}
	if page2Count != 2 {
		t.Errorf("expected 2 lines on page 2, got %d", page2Count)
	}
}

func TestBuildBlocks(t *testing.T) {
	lines := []TextLine{
		{Text: "行1", PageNo: 1},
		{Text: "行2", PageNo: 1},
		{Text: "", PageNo: 1},
		{Text: "行3", PageNo: 1},
		{Text: "", PageNo: 1},
		{Text: "", PageNo: 1},
		{Text: "行4", PageNo: 2},
	}
	blocks := BuildBlocks(lines)
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	if blocks[2].PageNo != 2 {
		t.Errorf("expected last block on page 2, got %d", blocks[2].PageNo)
	}
}

func TestMergeBrokenLinesInCleaner(t *testing.T) {
	c := NewCleaner()
	blocks := []TextBlock{
		{ID: "b1", PageNo: 1, Lines: []TextLine{{Text: "这是一个因为排版"}}},
		{ID: "b2", PageNo: 1, Lines: []TextLine{{Text: "原因被断开的句子"}}},
	}
	for i := range blocks {
		blocks[i].Text = joinLines(blocks[i].Lines)
	}
	result := c.Clean(blocks)
	if len(result) != 1 {
		t.Fatalf("expected 1 merged block, got %d", len(result))
	}
	if !strings.Contains(result[0].Text, "排版") || !strings.Contains(result[0].Text, "原因") {
		t.Errorf("merge failed, got: %s", result[0].Text)
	}
}
