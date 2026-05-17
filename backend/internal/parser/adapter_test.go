package parser

import "testing"

func TestStructuredDocumentAdapterPreservesBlockBoundary(t *testing.T) {
	adapter := NewStructuredDocumentAdapter()
	lines, err := adapter.Adapt(RawDocument{
		Pages: []RawPage{
			{
				PageNo: 1,
				Blocks: []RawBlock{
					{ID: "p1_b1", Lines: []RawLine{{Text: "给定资料一"}, {Text: "材料内容。"}}},
					{ID: "p1_b2", Lines: []RawLine{{Text: "第一题"}, {Text: "请概括问题。不超过300字。"}}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Adapt returned error: %v", err)
	}

	blocks := BuildBlocks(lines)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].ID != "p1_b1" || blocks[1].ID != "p1_b2" {
		t.Fatalf("block IDs not preserved: %+v", blocks)
	}
}
