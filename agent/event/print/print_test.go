package print

import (
	"strings"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestFormatAgentEventToolCallShowsBashCommand(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "bash",
		ToolInput: map[string]any{
			"command": "go test ./cmd/knowledge-hub",
		},
	})
	requireContains(t, got, "RUN")
	requireContains(t, got, "go test ./cmd/knowledge-hub")
}

func TestFormatAgentEventToolCallShowsReadPath(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "read",
		ToolInput: map[string]any{
			"filePath": "/tmp/work/src/main.go",
		},
	})
	requireContains(t, got, "READ")
	requireContains(t, got, "/tmp/work/src/main.go")
}

func TestFormatAgentEventToolCallShowsEditPath(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "edit",
		ToolInput: map[string]any{
			"filePath": "/tmp/work/src/main.go",
		},
	})
	requireContains(t, got, "EDIT")
	requireContains(t, got, "/tmp/work/src/main.go")
}

func TestFormatAgentEventToolCallShowsSkillName(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "skill",
		ToolInput: map[string]any{
			"name": "confluence-fetch",
		},
	})
	requireContains(t, got, "SKILL")
	requireContains(t, got, "confluence-fetch")
}

func TestFormatAgentEventToolCallShowsWebfetchURL(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "webfetch",
		ToolInput: map[string]any{
			"url": "https://example.test/page",
		},
	})
	requireContains(t, got, "WEBFETCH")
	requireContains(t, got, "https://example.test/page")
}

func TestFormatAgentEventToolCallShowsTodoDetails(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "todowrite",
		ToolInput: map[string]any{
			"todos": []any{
				map[string]any{"content": "Search docs", "status": "in_progress"},
			},
		},
	})
	requireContains(t, got, "TODO")
	requireContains(t, got, "in_progress: Search docs")
}

func TestFormatAgentEventToolCallShowsMCPToolDoc(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "skynet-base_get_doc_content",
		ToolInput: map[string]any{
			"doc": "https://confluence.example.test/search?q=pricing",
		},
	})
	requireContains(t, got, "SKYNET-BASE_GET_DOC_CONTENT")
	requireContains(t, got, "https://confluence.example.test/search?q=pricing")
}

func requireContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}
