package print

import (
	"fmt"
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

func TestFormatAgentEventToolCallShowsGlobPatternWithOutput(t *testing.T) {
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionToolCall,
		Tool: "glob",
		ToolInput: map[string]any{
			"pattern": "**/*.md",
		},
		Mock: &eventtypes.MockConfig{
			Output: "No files found",
		},
	})
	requireContains(t, got, "SEARCH")
	requireContains(t, got, "**/*.md")
	requireContains(t, got, "No files found")
	if strings.Index(got, "**/*.md") > strings.Index(got, "No files found") {
		t.Fatalf("pattern should appear before output:\n%s", got)
	}
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

func TestFormatAgentEventToolCallTruncatesLongLine(t *testing.T) {
	long := strings.Repeat("x", TRUNCATE_LINE_MAX+50)
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type:   eventtypes.ActionToolCall,
		Tool:   "bash",
		Output: long,
	})
	for _, line := range strings.Split(got, "\n") {
		if len([]rune(line)) > TRUNCATE_LINE_MAX+3 { // + "..."
			t.Fatalf("line exceeds cap (%d runes): %d\n%s", TRUNCATE_LINE_MAX+3, len([]rune(line)), line[:80])
		}
	}
	requireContains(t, got, "...")
}

func TestFormatAgentEventToolCallCapsAt16Lines(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "line-%d\n", i)
	}
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type:   eventtypes.ActionToolCall,
		Tool:   "bash",
		Output: b.String(),
	})
	lines := strings.Split(got, "\n")
	// header + up to 16 body/header lines total in truncateToolDisplay, then footer
	if len(lines) > TruncateToolMaxLines+1 {
		t.Fatalf("got %d lines, want ≤ %d:\n%s", len(lines), TruncateToolMaxLines+1, got)
	}
	requireContains(t, got, "lines truncated")
	if strings.Contains(got, "line-39") {
		t.Fatalf("late lines should be omitted:\n%s", got)
	}
}

func TestFormatAgentEventMessageNotLineCapped(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "assist-%d\n", i)
	}
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionMessage,
		Text: b.String(),
	})
	if strings.Contains(got, "truncated") {
		t.Fatalf("message must not be line-capped:\n%s", got)
	}
	requireContains(t, got, "assist-39")
}

func TestFormatAgentEventThinkNotLineCapped(t *testing.T) {
	longLine := strings.Repeat("t", TRUNCATE_LINE_MAX+100)
	got := FormatAgentEvent(eventtypes.AgentEvent{
		Type: eventtypes.ActionThink,
		Text: longLine,
	})
	if strings.Contains(got, "...") && !strings.Contains(got, longLine) {
		t.Fatalf("think must not truncate body:\n%s", got[:120])
	}
	if !strings.Contains(got, longLine) {
		t.Fatalf("think text missing")
	}
}

func TestFormatAgentEventForStdoutToolTruncates(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&b, "out-%d\n", i)
	}
	got := FormatAgentEventForStdout(eventtypes.AgentEvent{
		Type:   eventtypes.ActionToolCall,
		Tool:   "bash",
		Output: b.String(),
	})
	requireContains(t, got, "truncated")
}

func requireContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}
