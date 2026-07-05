# Scenario

**Feature**: grok session updates.jsonl ↔ canonical AgentEvent conversion

```
# forward: ACP wire lines coalesce into canonical events with grok_session extensions
updates.jsonl wire lines -> ParseLine / Converter -> []types.AgentEvent

# reverse: canonical events split/re-chunk into session updates
[]types.AgentEvent -> ToSession -> ToWireLines -> updates.jsonl wire

# compatibility proof: semantic equality across wire → events → wire → events
wire₁ -> events₁ -> wire₂ -> events₂; SemanticEqual(events₁, events₂)
```

## Preconditions

- The `grok_session` package must exist under `agent/event/grok_session`.
- `types.AgentEvent` carries `tool_call_id` and `extensions.grok_session` (`status`, `turn_index`).
- Synthetic ACP wire lines are built by helpers in this file (no real grok binary).
- `SemanticEqual` compares canonical events ignoring timestamps and wire chunk fragmentation.

## Steps

1. For `Target="from_session"`: call `grok_session.FromUpdatesJSONL(req.WireLines)` and marshal events.
2. For `Target="to_session"`: call `grok_session.ToSession` then `grok_session.ToWireLines`.
3. For `Target="roundtrip"`: run `wire₁ → events₁ → wire₂ → events₂` and expose both event slices.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = "sess_grok_test_001"
	}
	return nil
}

// --- ACP flat wire line builders ---

func acpUserChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpThoughtChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_thought_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpAssistantChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpToolCall(toolCallID, kind, title string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call",
		"toolCallId":    toolCallID,
		"kind":          kind,
		"title":         title,
		"status":        "pending",
	})
	return string(line)
}

func acpToolCallUpdate(toolCallID, status, output string) string {
	content := []map[string]any{}
	if output != "" {
		content = append(content, map[string]any{
			"type": "content",
			"content": map[string]any{
				"type": "text",
				"text": output,
			},
		})
	}
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call_update",
		"toolCallId":    toolCallID,
		"status":        status,
		"content":       content,
	})
	return string(line)
}

func acpTurnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

// acpEnvelope wraps a flat update JSON object in the grok wire envelope.
func acpEnvelope(sessionID string, flatUpdate map[string]any) string {
	inner, _ := json.Marshal(flatUpdate)
	payload := map[string]any{
		"method": "_x.ai/session/update",
		"params": map[string]any{
			"sessionId": sessionID,
			"update":    json.RawMessage(inner),
		},
	}
	line, _ := json.Marshal(payload)
	return string(line)
}

func acpEnvelopeUserChunk(sessionID, text string) string {
	return acpEnvelope(sessionID, map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
}

// --- assertion helpers ---

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}

func grokTurnIndex(ev types.AgentEvent) int {
	if ev.Extensions == nil || ev.Extensions.GrokSession == nil {
		return -1
	}
	return ev.Extensions.GrokSession.TurnIndex
}

func grokStatus(ev types.AgentEvent) string {
	if ev.Extensions == nil || ev.Extensions.GrokSession == nil {
		return ""
	}
	return ev.Extensions.GrokSession.Status
}

// semanticEvent projects an AgentEvent to comparable fields (ignores timestamp/id).
type semanticEvent struct {
	Type       types.ActionType
	Role       string
	Text       string
	Tool       string
	Output     string
	ToolCallID string
	Status     string
	TurnIndex  int
}

func toSemantic(ev types.AgentEvent) semanticEvent {
	return semanticEvent{
		Type:       ev.Type,
		Role:       ev.Role,
		Text:       ev.Text,
		Tool:       ev.Tool,
		Output:     ev.Output,
		ToolCallID: ev.ToolCallID,
		Status:     grokStatus(ev),
		TurnIndex:  grokTurnIndex(ev),
	}
}

// SemanticEqual reports whether two event slices match on semantic fields.
func SemanticEqual(got, want []types.AgentEvent) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if toSemantic(got[i]) != toSemantic(want[i]) {
			return false
		}
	}
	return true
}

func assertSemanticEqual(t *testing.T, got, want []types.AgentEvent) {
	t.Helper()
	if SemanticEqual(got, want) {
		return
	}
	gb, _ := json.MarshalIndent(got, "", "  ")
	wb, _ := json.MarshalIndent(want, "", "  ")
	t.Fatalf("semantic events mismatch:\ngot:\n%s\nwant:\n%s", gb, wb)
}

func assertSemanticEqualEvents(t *testing.T, events1, events2 []types.AgentEvent) {
	t.Helper()
	assertSemanticEqual(t, events1, events2)
}

func assertEventCount(t *testing.T, events []types.AgentEvent, want int) {
	t.Helper()
	if len(events) != want {
		t.Fatalf("event count: got %d want %d\n%#v", len(events), want, events)
	}
}

func assertAllTurnIndex(t *testing.T, events []types.AgentEvent, want int) {
	t.Helper()
	for i, ev := range events {
		if got := grokTurnIndex(ev); got != want {
			t.Fatalf("events[%d] turn_index: got %d want %d (%#v)", i, got, want, ev)
		}
	}
}

func assertHasTurnIndex(t *testing.T, events []types.AgentEvent, indices ...int) {
	t.Helper()
	if len(indices) == 0 {
		return
	}
	seen := make(map[int]bool)
	for _, ev := range events {
		seen[grokTurnIndex(ev)] = true
	}
	for _, want := range indices {
		if !seen[want] {
			t.Fatalf("missing turn_index %d in events:\n%#v", want, events)
		}
	}
}

func assertWireHasSessionUpdate(t *testing.T, wire []string, update string) {
	t.Helper()
	for _, line := range wire {
		if strings.Contains(line, `"sessionUpdate":`+jsonQuote(update)) ||
			strings.Contains(line, `"sessionUpdate":"`+update+`"`) {
			return
		}
	}
	joined, _ := json.Marshal(wire)
	t.Fatalf("wire missing sessionUpdate %q:\n%s", update, joined)
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func assertWireToolCallID(t *testing.T, wire []string, id string) {
	t.Helper()
	for _, line := range wire {
		if strings.Contains(line, `"toolCallId":"`+id+`"`) {
			return
		}
	}
	joined, _ := json.Marshal(wire)
	t.Fatalf("wire missing toolCallId %q:\n%s", id, joined)
}

func assertEventsOfType(t *testing.T, events []types.AgentEvent, typ types.ActionType, want int) {
	t.Helper()
	n := 0
	for _, ev := range events {
		if ev.Type == typ {
			n++
		}
	}
	if n != want {
		t.Fatalf("type %s count: got %d want %d\n%#v", typ, n, want, events)
	}
}

func formatEvents(events []types.AgentEvent) string {
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", events)
	}
	return string(b)
}
```