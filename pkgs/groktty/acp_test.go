package groktty

import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestACPConverterFullSequence(t *testing.T) {
	c := NewACPConverter()
	lines := []string{
		acpLine("user_message_chunk", "persist events"),
		acpLine("agent_thought_chunk", "planning ls output"),
		acpLine("tool_call", ""),
		acpLine("tool_call_update", ""),
		acpLine("agent_message_chunk", "PERSIST_ASSISTANT_MARKER"),
	}
	var events []types.AgentEvent
	for _, line := range lines {
		events = append(events, c.ProcessLine(line)...)
	}
	events = append(events, c.Flush()...)

	found := map[string]bool{}
	for _, ev := range events {
		switch ev.Type {
		case types.ActionMessage:
			if ev.Role == "user" {
				found["user"] = true
			}
			if ev.Role == "assistant" {
				found["assistant"] = true
			}
		case types.ActionThink:
			found["think"] = true
		case types.ActionToolCall:
			found["tool"] = true
		}
	}
	for _, key := range []string{"user", "think", "tool", "assistant"} {
		if !found[key] {
			t.Fatalf("missing %s event in %#v", key, events)
		}
	}
}

func TestACPConverterNestedWireFormat(t *testing.T) {
	c := NewACPConverter()
	line := `{"method":"session/update","params":{"sessionId":"550e8400-e29b-41d4-a716-446655440000","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"run ls"}}}}`
	events := c.ProcessLine(line)
	events = append(events, c.Flush()...)
	foundUser := false
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "user" && ev.Text == "run ls" {
			foundUser = true
		}
	}
	if !foundUser {
		t.Fatalf("expected user event from nested wire line, got %#v", events)
	}
}

func TestACPConverterRealGrokWireLine(t *testing.T) {
	c := NewACPConverter()
	line := `{"timestamp":1782714542,"method":"_x.ai/session/update","params":{"sessionId":"019f1211-106c-77f0-9726-95e88090aac7","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"agent"}}}}`
	events := c.ProcessLine(line)
	events = append(events, c.Flush()...)
	found := false
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "assistant" && ev.Text == "agent" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected assistant from _x.ai/session/update wire line, got %#v", events)
	}
}

func TestParseACPUpdateLineNestedWire(t *testing.T) {
	line := `{"method":"session/update","params":{"sessionId":"abc","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"hello"}}}}`
	upd, ok := parseACPUpdateLine(line)
	if !ok {
		t.Fatal("parseACPUpdateLine returned false for nested wire")
	}
	if upd.SessionUpdate != "agent_message_chunk" {
		t.Fatalf("sessionUpdate=%q", upd.SessionUpdate)
	}
	if got := acpTextContent(upd.Content); got != "hello" {
		t.Fatalf("content=%q", got)
	}
}

func acpLine(kind, text string) string {
	switch kind {
	case "user_message_chunk", "agent_thought_chunk", "agent_message_chunk":
		return `{"sessionUpdate":"` + kind + `","content":{"type":"text","text":"` + text + `"}}`
	case "tool_call":
		return `{"sessionUpdate":"tool_call","toolCallId":"call_persist","kind":"execute","title":"ls","status":"pending"}`
	case "tool_call_update":
		return `{"sessionUpdate":"tool_call_update","toolCallId":"call_persist","status":"completed","content":[{"type":"content","content":{"type":"text","text":"agent\nagents"}}]}`
	default:
		return ""
	}
}