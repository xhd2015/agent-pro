package codex

import (
	"testing"

	"github.com/xhd2015/agent-traces/agent/cli/registry"
)

func TestExtractFileChanges(t *testing.T) {
	raw := map[string]any{
		"changes": []any{
			map[string]any{"path": "/tmp/a.md", "kind": "add"},
			map[string]any{"path": "/tmp/b.md", "kind": "updated"},
			map[string]any{"path": "/tmp/c.md", "kind": "delete"},
		},
	}

	got := extractFileChanges(raw)
	want := []registry.FileChange{
		{Path: "/tmp/a.md", Kind: "add"},
		{Path: "/tmp/b.md", Kind: "modify"},
		{Path: "/tmp/c.md", Kind: "delete"},
	}

	if len(got) != len(want) {
		t.Fatalf("len(file changes) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("file change %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestExtractToolCallEventForFileChange(t *testing.T) {
	event := codexEvent{
		Type: "item.completed",
		Item: &codexItem{
			ID:     "item_43",
			Type:   "file_change",
			Status: "completed",
			Raw: map[string]any{
				"changes": []any{
					map[string]any{"path": "/repo/docs/overview.md", "kind": "add"},
					map[string]any{"path": "/repo/docs/faq.md", "kind": "modify"},
				},
			},
		},
	}

	got := event.extractToolCallEvent()
	if got == nil {
		t.Fatal("extractToolCallEvent returned nil")
	}
	if got.ToolName != "File Change" {
		t.Fatalf("tool name = %q, want %q", got.ToolName, "File Change")
	}
	if got.Summary != "2 files changed\n+ /repo/docs/overview.md\n~ /repo/docs/faq.md" {
		t.Fatalf("summary = %q", got.Summary)
	}
	if got.Status != "completed" {
		t.Fatalf("status = %q, want completed", got.Status)
	}
	if !got.ReplaceSummary {
		t.Fatal("replace summary = false, want true")
	}
	if len(got.FileChanges) != 2 {
		t.Fatalf("len(file_changes) = %d, want 2", len(got.FileChanges))
	}
}

func TestExtractToolCallEventForFailedMCPToolCall(t *testing.T) {
	event := codexEvent{
		Type: "item.completed",
		Item: &codexItem{
			ID:     "item_35",
			Type:   "mcp_tool_call",
			Status: "failed",
			Raw: map[string]any{
				"server": "skynet-base",
				"tool":   "get_doc_content",
				"arguments": map[string]any{
					"doc": "https://docs.google.com/spreadsheets/d/example/edit#gid=0",
				},
				"result": map[string]any{
					"content": []any{
						map[string]any{
							"type": "text",
							"text": "Error: Authentication required. Please run 'skynet auth login' command and try again.",
						},
					},
				},
			},
		},
	}

	got := event.extractToolCallEvent()
	if got == nil {
		t.Fatal("extractToolCallEvent returned nil")
	}
	wantSummary := "skynet-base.get_doc_content\ndoc: https://docs.google.com/spreadsheets/d/example/edit#gid=0\nError: Authentication required. Please run 'skynet auth login' command and try again."
	if got.Summary != wantSummary {
		t.Fatalf("summary = %q, want %q", got.Summary, wantSummary)
	}
	if got.Status != "failed" {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	if !got.ReplaceSummary {
		t.Fatal("replace summary = false, want true")
	}
}
