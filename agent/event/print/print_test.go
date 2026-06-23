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

func requireContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}
