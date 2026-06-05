package fakeagent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatCodexEvents_Valid(t *testing.T) {
	exitCode := 0
	events := []Event{
		{
			Type: EventStarted,
			Item: &EventItem{
				ID:   "item_1",
				Type: ItemReasoning,
			},
		},
		{
			Type: EventCompleted,
			Item: &EventItem{
				ID:     "item_1",
				Type:   ItemReasoning,
				Text:   "Let me think about this.",
				Status: "completed",
			},
		},
		{
			Type: EventStarted,
			Item: &EventItem{
				ID:      "item_2",
				Type:    ItemCommandExecution,
				Command: "ls -la",
			},
		},
		{
			Type: EventCompleted,
			Item: &EventItem{
				ID:               "item_2",
				Type:             ItemCommandExecution,
				Command:          "ls -la",
				AggregatedOutput: "total 4\ndrwxr-xr-x\n",
				ExitCode:         &exitCode,
				Status:           "completed",
			},
		},
		{
			Type: EventCompleted,
			Item: &EventItem{
				ID:     "item_3",
				Type:   ItemMessage,
				Text:   "Done!",
				Status: "completed",
			},
		},
	}

	lines, err := FormatCodexEvents(events)
	if err != nil {
		t.Fatalf("FormatCodexEvents failed: %v", err)
	}

	if len(lines) != len(events) {
		t.Fatalf("got %d lines, want %d", len(lines), len(events))
	}

	for i, line := range lines {
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("line %d is not valid JSON: %v\nline: %s", i, err, line)
		}
		if event.Type == "" {
			t.Fatalf("line %d: event has empty type", i)
		}
	}
}

func TestParseCodexEvents_RoundTrip(t *testing.T) {
	g := NewGenerator(42)
	events := g.GenerateSession("write a test")

	lines, err := FormatCodexEvents(events)
	if err != nil {
		t.Fatalf("FormatCodexEvents failed: %v", err)
	}

	parsed, err := ParseCodexEvents(lines)
	if err != nil {
		t.Fatalf("ParseCodexEvents failed: %v", err)
	}

	if len(parsed) != len(events) {
		t.Fatalf("round-trip: got %d events, want %d", len(parsed), len(events))
	}

	for i := range events {
		if parsed[i].Type != events[i].Type {
			t.Fatalf("event %d: type mismatch: %s vs %s", i, parsed[i].Type, events[i].Type)
		}
	}
}

func TestParseCodexEvents_InvalidJSON(t *testing.T) {
	_, err := ParseCodexEvents([]string{"not json"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFormatCodexEvents_Empty(t *testing.T) {
	lines, err := FormatCodexEvents(nil)
	if err != nil {
		t.Fatalf("FormatCodexEvents(nil) failed: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines for nil events, got %d", len(lines))
	}

	lines, err = FormatCodexEvents([]Event{})
	if err != nil {
		t.Fatalf("FormatCodexEvents([]) failed: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines for empty events, got %d", len(lines))
	}
}

func TestCodexEventJSONMatchesCodexSchema(t *testing.T) {
	exitCode := 0
	event := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               "item_cmd_1",
			Type:             ItemCommandExecution,
			Command:          "go test ./...",
			AggregatedOutput: "ok",
			ExitCode:         &exitCode,
			Status:           "completed",
		},
	}

	b, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed["type"] != "item.completed" {
		t.Fatalf("type = %v, want item.completed", parsed["type"])
	}

	item, ok := parsed["item"].(map[string]any)
	if !ok {
		t.Fatal("item field missing or not an object")
	}

	if item["command"] != "go test ./..." {
		t.Fatalf("command = %v, want go test ./...", item["command"])
	}

	if item["exit_code"] != float64(0) {
		t.Fatalf("exit_code = %v, want 0", item["exit_code"])
	}
}

func TestFileChangeJSON(t *testing.T) {
	event := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:   "item_fc_1",
			Type: ItemFileChange,
			Status: "completed",
			Changes: []FileChange{
				{Path: "/tmp/a.go", Kind: "add"},
				{Path: "/tmp/b.go", Kind: "modify"},
			},
		},
	}

	b, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	item := parsed["item"].(map[string]any)
	changes, ok := item["changes"].([]any)
	if !ok {
		t.Fatal("changes not an array")
	}
	if len(changes) != 2 {
		t.Fatalf("got %d changes, want 2", len(changes))
	}
}

func TestFormatCodexEventsText_Reasoning(t *testing.T) {
	events := []Event{
		{Type: EventStarted, Item: &EventItem{ID: "r1", Type: ItemReasoning}},
		{Type: EventUpdated, Item: &EventItem{ID: "r1", Type: ItemReasoning, Text: "Let me check"}},
		{Type: EventCompleted, Item: &EventItem{ID: "r1", Type: ItemReasoning, Text: "Let me think.\nI will plan.", Status: "completed"}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "Thinking...") {
		t.Fatalf("expected 'Thinking...', got: %s", text)
	}
	if !strings.Contains(text, "Let me think.") {
		t.Fatalf("expected reasoning text: %s", text)
	}
	if !strings.Contains(text, "I will plan.") {
		t.Fatalf("expected second line: %s", text)
	}
}

func TestFormatCodexEventsText_CommandExecution(t *testing.T) {
	exitCode := 0
	events := []Event{
		{Type: EventStarted, Item: &EventItem{ID: "c1", Type: ItemCommandExecution, Command: "ls -la"}},
		{Type: EventCompleted, Item: &EventItem{
			ID: "c1", Type: ItemCommandExecution, Command: "ls -la",
			AggregatedOutput: "file1.txt\nfile2.txt", ExitCode: &exitCode, Status: "completed",
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "> ls -la") {
		t.Fatalf("expected '> ls -la': %s", text)
	}
	if !strings.Contains(text, "file1.txt") {
		t.Fatalf("expected file1.txt: %s", text)
	}
	if !strings.Contains(text, "file2.txt") {
		t.Fatalf("expected file2.txt: %s", text)
	}
}

func TestFormatCodexEventsText_FileChangeAdd(t *testing.T) {
	events := []Event{
		{Type: EventCompleted, Item: &EventItem{
			ID: "f1", Type: ItemFileChange, Status: "completed",
			Changes: []FileChange{{Path: "/tmp/new.go", Kind: "add"}},
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "+ /tmp/new.go (created)") {
		t.Fatalf("expected file add: %s", text)
	}
}

func TestFormatCodexEventsText_FileChangeModify(t *testing.T) {
	events := []Event{
		{Type: EventCompleted, Item: &EventItem{
			ID: "f1", Type: ItemFileChange, Status: "completed",
			Changes: []FileChange{{Path: "/tmp/edit.go", Kind: "modify"}},
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "~ /tmp/edit.go (modified)") {
		t.Fatalf("expected file modify: %s", text)
	}
}

func TestFormatCodexEventsText_FileChangeDelete(t *testing.T) {
	events := []Event{
		{Type: EventCompleted, Item: &EventItem{
			ID: "f1", Type: ItemFileChange, Status: "completed",
			Changes: []FileChange{{Path: "/tmp/old.go", Kind: "delete"}},
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "- /tmp/old.go (deleted)") {
		t.Fatalf("expected file delete: %s", text)
	}
}

func TestFormatCodexEventsText_Message(t *testing.T) {
	events := []Event{
		{Type: EventCompleted, Item: &EventItem{
			ID: "m1", Type: ItemMessage, Text: "All done!", Status: "completed",
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "All done!") {
		t.Fatalf("expected message: %s", text)
	}
}

func TestFormatCodexEventsText_SkipsStarted(t *testing.T) {
	events := []Event{
		{Type: EventStarted, Item: &EventItem{ID: "c1", Type: ItemCommandExecution, Command: "skipped"}},
	}
	text := FormatCodexEventsText(events)
	if strings.Contains(text, "skipped") {
		t.Fatalf("started event should be skipped: %s", text)
	}
}

func TestFormatCodexEventsText_SkipsUpdated(t *testing.T) {
	events := []Event{
		{Type: EventUpdated, Item: &EventItem{ID: "r1", Type: ItemReasoning, Text: "partial thought"}},
	}
	text := FormatCodexEventsText(events)
	if strings.Contains(text, "partial thought") {
		t.Fatalf("updated event should be skipped: %s", text)
	}
}

func TestFormatCodexEventsText_EmptyEvents(t *testing.T) {
	text := FormatCodexEventsText(nil)
	if text != "" {
		t.Fatalf("expected empty for nil, got: %s", text)
	}
	text = FormatCodexEventsText([]Event{})
	if text != "" {
		t.Fatalf("expected empty for empty slice, got: %s", text)
	}
}

func TestFormatCodexEventsText_CommandNoOutput(t *testing.T) {
	exitCode := 0
	events := []Event{
		{Type: EventCompleted, Item: &EventItem{
			ID: "c1", Type: ItemCommandExecution, Command: "echo done",
			ExitCode: &exitCode, Status: "completed",
		}},
	}
	text := FormatCodexEventsText(events)
	if !strings.Contains(text, "> echo done") {
		t.Fatalf("expected command line: %s", text)
	}
}
