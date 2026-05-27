package agent_trace

import (
	"strings"
	"testing"
)

func TestAgentTraceParsesPlanAndFileChangeDetails(t *testing.T) {
	lines := []string{
		`{"type":"item.started","item":{"id":"item_7","type":"todo_list","items":[{"text":"Inspect Jira comments","completed":false},{"text":"Write output JSON","completed":false}]}}`,
		`{"type":"item.updated","item":{"id":"item_7","type":"todo_list","items":[{"text":"Inspect Jira comments","completed":true},{"text":"Write output JSON","completed":false}]}}`,
		`{"type":"item.completed","item":{"id":"item_8","type":"file_change","changes":[{"path":"/tmp/code-commits.json","kind":"add"}],"status":"completed"}}`,
	}

	messages := ParseMessages(lines, "2026-04-28T16:57:57.816512+08:00")
	if len(messages) != 2 {
		t.Fatalf("messages = %d, want 2: %#v", len(messages), messages)
	}
	plan := messages[0].ToolCall
	if plan == nil || plan.ToolName != "Plan" {
		t.Fatalf("first message did not parse plan: %#v", messages[0])
	}
	if !strings.Contains(plan.Summary, "[x] Inspect Jira comments") || !strings.Contains(plan.Summary, "[ ] Write output JSON") {
		t.Fatalf("plan summary missing todo details: %q", plan.Summary)
	}

	fileChange := messages[1].ToolCall
	if fileChange == nil || fileChange.ToolName != "File Change" {
		t.Fatalf("second message did not parse file change: %#v", messages[1])
	}
	if len(fileChange.FileChanges) != 1 || fileChange.FileChanges[0].Path != "/tmp/code-commits.json" || fileChange.FileChanges[0].Kind != "add" {
		t.Fatalf("file changes mismatch: %#v", fileChange.FileChanges)
	}
	if !strings.Contains(fileChange.Summary, "+ /tmp/code-commits.json") {
		t.Fatalf("file change summary missing path: %q", fileChange.Summary)
	}
}

func TestAgentTraceParsesCodexHooksDeprecationAsWarning(t *testing.T) {
	const warning = "`[features].codex_hooks` is deprecated. Use `[features].hooks` instead."
	lines := []string{
		`{"type":"item.completed","item":{"id":"item_0","type":"error","message":"` + warning + `"}}`,
	}

	messages := ParseMessages(lines, "2026-05-25T18:26:22.524536+08:00")
	if len(messages) != 1 {
		t.Fatalf("messages = %d, want 1: %#v", len(messages), messages)
	}
	toolCall := messages[0].ToolCall
	if toolCall == nil {
		t.Fatalf("message did not parse as tool call: %#v", messages[0])
	}
	if toolCall.ToolName != "Config Warning" {
		t.Fatalf("tool name = %q, want Config Warning", toolCall.ToolName)
	}
	if toolCall.Kind != "warning" {
		t.Fatalf("kind = %q, want warning", toolCall.Kind)
	}
	if toolCall.Status != "warning" {
		t.Fatalf("status = %q, want warning", toolCall.Status)
	}
	if toolCall.Summary != warning {
		t.Fatalf("summary = %q, want %q", toolCall.Summary, warning)
	}
}

func TestAgentTraceParsesCursorToolCallViaAdapter(t *testing.T) {
	lines := []string{
		`{"type":"tool_call","subtype":"started","call_id":"cursor_1","tool_call":{"shellToolCall":{"args":{"command":"go test ./..."}}}}`,
		`{"type":"tool_call","subtype":"completed","call_id":"cursor_1","tool_call":{"shellToolCall":{"result":{"exit_code":0,"output":"ok"}}}}`,
	}

	messages := ParseMessages(lines, "2026-05-25T18:26:22.524536+08:00")
	if len(messages) != 1 {
		t.Fatalf("messages = %d, want 1: %#v", len(messages), messages)
	}
	toolCall := messages[0].ToolCall
	if toolCall == nil {
		t.Fatalf("message did not parse as tool call: %#v", messages[0])
	}
	if toolCall.CallID != "cursor_1" {
		t.Fatalf("call id = %q, want cursor_1", toolCall.CallID)
	}
	if toolCall.ToolName != "Shell" {
		t.Fatalf("tool name = %q, want Shell", toolCall.ToolName)
	}
	if toolCall.Status != "completed" {
		t.Fatalf("status = %q, want completed", toolCall.Status)
	}
	if messages[0].FinishedAt == nil {
		t.Fatalf("finished_at is nil for completed cursor tool call: %#v", messages[0])
	}
}
