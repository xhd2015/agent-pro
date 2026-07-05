# Scenario

**Feature**: groktty delegates ACP conversion to grok_session

```
# tail path: poll updates.jsonl and emit grok_session-rich AgentEvents
updates.jsonl -> TailUpdatesFromOffset -> emit(AgentEvent)

# session path: discover grok session dir by cwd + first user chunk
summary.json + updates.jsonl -> DiscoverSession -> session id + updates path

# parity: tail output must match grok_session.FromUpdatesJSONL semantics
wire lines -> tail events ≡ grok_session.FromUpdatesJSONL(wire lines)
```

## Preconditions

- `pkgs/groktty` exposes `TailUpdates`, `TailUpdatesFromOffset`, and `DiscoverSession`.
- After refactor, tail and session paths use `grok_session` (not local `acp.go`).
- Synthetic ACP wire lines are built by helpers below (patterns from
  `agent/event/grok_session/tests/SETUP.md` and `cmd/agent-run/tests/grok-tty/SETUP.md`).
- No real grok, no PTY — temp `updates.jsonl` and session dirs only.

## Steps

1. Root `Setup` initializes `SessionID`, `RunStart`, and shared defaults.
2. Grouping `Setup` sets `req.Target` (`tail`, `session`, or `integration`).
3. Leaf `Setup` writes wire lines or seeds a grok session dir.
4. `Run` tails updates or calls `DiscoverSession`.
5. Leaf `Assert` checks grok_session fields or semantic parity.

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

func tailAllEvents(t *testing.T, path string) ([]types.AgentEvent, error) {
	t.Helper()
	var events []types.AgentEvent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := groktty.TailUpdatesFromOffset(ctx, path, 0, func(ev types.AgentEvent) error {
		events = append(events, ev)
		return nil
	})
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		return events, err
	}
	return events, nil
}

func writeUpdatesJSONL(path string, lines ...string) error {
	var b strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = "sess_grok_delegate_001"
	}
	if req.RunStart.IsZero() {
		req.RunStart = time.Now().Add(-time.Second)
	}
	return nil
}

// --- ACP flat wire line builders (grok-tty / grok_session patterns) ---

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

func acpEnvelopeAssistantChunk(sessionID, text string) string {
	return acpEnvelope(sessionID, map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
}

// --- session dir seeding ---

func seedGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, created time.Time, updatesLines ...string) string {
	t.Helper()
	encoded, err := groktty.EncodedCwd(workspace)
	if err != nil {
		t.Fatalf("EncodedCwd: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", encoded, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	abs, _ := filepath.Abs(workspace)
	summary := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
		},
		"created_at": created.UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	if len(updatesLines) == 0 {
		updatesLines = []string{acpUserChunk(prompt)}
	}
	if err := writeUpdatesJSONL(updatesPath, updatesLines...); err != nil {
		t.Fatalf("write updates.jsonl: %v", err)
	}
	return updatesPath
}

func newTempUpdatesPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "updates.jsonl")
}

// --- assertion helpers ---

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

func assertAllTurnIndex(t *testing.T, events []types.AgentEvent, want int) {
	t.Helper()
	for i, ev := range events {
		if got := grokTurnIndex(ev); got != want {
			t.Fatalf("events[%d] turn_index: got %d want %d (%#v)", i, got, want, ev)
		}
	}
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

func toolEvents(events []types.AgentEvent) []types.AgentEvent {
	var out []types.AgentEvent
	for _, ev := range events {
		if ev.Type == types.ActionToolCall {
			out = append(out, ev)
		}
	}
	return out
}

func formatEvents(events []types.AgentEvent) string {
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", events)
	}
	return string(b)
}
```