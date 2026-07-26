---
label: slow, heavy
explanation: runs the real claude binary headless; LLM latency + cost
---

## Expected
- `Response.Answer` contains "pong" (case-insensitive).
- `Response.SessionID` is non-empty (session id captured from the `system` init
  and/or `result` event).
- `Response.Events` is non-empty.
- Every entry in `Response.Events` is valid JSON.
- At least one AgentEvent has `"type":"message"` (assistant text written to RawLog).
- No error occurred.

## Side Effects
- Spawns the real `claude` binary; consumes LLM tokens.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected non-empty SessionID after Ask()")
	}
	lower := strings.ToLower(resp.Answer)
	if !strings.Contains(lower, "pong") {
		t.Fatalf("expected answer to contain 'pong', got: %s", resp.Answer)
	}
	if len(resp.Events) == 0 {
		t.Fatal("expected non-empty Events after Ask()")
	}
	hasMessage := false
	for _, raw := range resp.Events {
		if !json.Valid(raw) {
			t.Fatalf("invalid JSON event: %s", string(raw))
		}
		var ev map[string]any
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if typ, _ := ev["type"].(string); typ == "message" {
			hasMessage = true
		}
	}
	if !hasMessage {
		t.Fatal("expected at least one AgentEvent with type 'message' in RawLog")
	}
}
```
