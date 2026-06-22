package grok

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestWriteAgentEventsFromGrokLine_Text(t *testing.T) {
	var buf bytes.Buffer
	writeAgentEventsFromGrokLine(&buf, `{"type":"text","data":"hello"}`)

	var event eventtypes.AgentEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.Type != eventtypes.ActionMessage {
		t.Fatalf("type = %q, want message", event.Type)
	}
	if event.Text != "hello" {
		t.Fatalf("text = %q, want hello", event.Text)
	}
}

func TestWriteAgentEventsFromGrokLine_Thought(t *testing.T) {
	var buf bytes.Buffer
	writeAgentEventsFromGrokLine(&buf, `{"type":"thought","data":"planning"}`)

	var event eventtypes.AgentEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.Type != eventtypes.ActionThink {
		t.Fatalf("type = %q, want think", event.Type)
	}
	if event.Text != "planning" {
		t.Fatalf("text = %q, want planning", event.Text)
	}
}

func TestWriteAgentEventsFromGrokLine_End(t *testing.T) {
	var buf bytes.Buffer
	writeAgentEventsFromGrokLine(&buf, `{"type":"end","sessionId":"sess-123"}`)

	var event eventtypes.AgentEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.Type != eventtypes.ActionDone {
		t.Fatalf("type = %q, want done", event.Type)
	}
	if event.ToolInput["session_id"] != "sess-123" {
		t.Fatalf("session_id = %v, want sess-123", event.ToolInput["session_id"])
	}
}

func TestWriteAgentEventsFromGrokLine_SkipsEmptyDelta(t *testing.T) {
	var buf bytes.Buffer
	writeAgentEventsFromGrokLine(&buf, `{"type":"text","data":""}`)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for empty text delta, got %q", buf.String())
	}
}

func TestWriteAgentEventsFromGrokLine_SkipsNonJSON(t *testing.T) {
	var buf bytes.Buffer
	writeAgentEventsFromGrokLine(&buf, "not json")
	if buf.Len() != 0 {
		t.Fatalf("expected no output for non-json line, got %q", buf.String())
	}
}

func TestWriteAgentEventsFromGrokLine_NilWriter(t *testing.T) {
	writeAgentEventsFromGrokLine(nil, `{"type":"text","data":"hello"}`)
}

func TestWriteAgentEventsFromGrokLine_StreamedThoughtDeltas(t *testing.T) {
	input := []string{
		`{"type":"thought","data":"The"}`,
		`{"type":"thought","data":" user"}`,
		`{"type":"thought","data":" acts"}`,
	}
	var buf bytes.Buffer
	w := NewGrokEventWriter(&buf)
	for _, line := range input {
		w.WriteGrokLine(line)
	}
	w.Flush()

	converted := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(converted) != 1 {
		t.Fatalf("got %d converted lines, want 1", len(converted))
	}
	var ev eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(converted[0]), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != eventtypes.ActionThink || ev.Text != "The user acts" {
		t.Fatalf("event = %+v, want think 'The user acts'", ev)
	}
}

func TestWriteAgentEventsFromGrokLine_StreamedTextDeltas(t *testing.T) {
	input := []string{
		`{"type":"text","data":"Hel"}`,
		`{"type":"text","data":"lo"}`,
	}
	var buf bytes.Buffer
	for _, line := range input {
		writeAgentEventsFromGrokLine(&buf, line)
	}

	converted := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(converted) != 2 {
		t.Fatalf("got %d converted lines, want 2", len(converted))
	}

	var first, second eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(converted[0]), &first); err != nil {
		t.Fatalf("unmarshal first: %v", err)
	}
	if err := json.Unmarshal([]byte(converted[1]), &second); err != nil {
		t.Fatalf("unmarshal second: %v", err)
	}
	if first.Type != eventtypes.ActionMessage || first.Text != "Hel" {
		t.Fatalf("first event = %+v, want message Hel", first)
	}
	if second.Type != eventtypes.ActionMessage || second.Text != "lo" {
		t.Fatalf("second event = %+v, want message lo", second)
	}
}
