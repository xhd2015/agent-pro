## Expected
- Response.Answer contains "paris" (case-insensitive).
- Response.SessionID is non-empty (session ID captured from `end` event).
- Response.Events is non-empty.
- Every entry in Response.Events is valid JSON.
- At least one event has `"type":"message"` (AgentEvent written to RawLog).
- At least one event has `"type":"done"` with session ID in `tool_input`, or `SessionID` is set from the grok `end` line.

## Side Effects
- None.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected non-empty SessionID after Ask()")
	}
	lower := strings.ToLower(resp.Answer)
	if !strings.Contains(lower, "paris") {
		t.Fatalf("expected answer to contain 'paris', got: %s", resp.Answer)
	}
	if len(resp.Events) == 0 {
		t.Fatal("expected non-empty Events after Ask()")
	}
	hasMessage := false
	hasDone := false
	for _, raw := range resp.Events {
		if !json.Valid(raw) {
			t.Fatalf("invalid JSON event: %s", string(raw))
		}
		var ev map[string]any
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		typ, _ := ev["type"].(string)
		if typ == "message" {
			hasMessage = true
		}
		if typ == "done" {
			if toolInput, ok := ev["tool_input"].(map[string]any); ok {
				if sid, _ := toolInput["session_id"].(string); sid != "" {
					hasDone = true
				}
			}
		}
	}
	if !hasMessage {
		t.Fatal("expected at least one AgentEvent with type 'message' in RawLog")
	}
	if !hasDone && resp.SessionID == "" {
		t.Fatal("expected done event with session_id in tool_input or non-empty SessionID")
	}
}
```
