## Expected
- Response.Answer contains "paris" (case-insensitive).
- Response.SessionID is non-empty (a session was created by the pi server).
- Response.Events is non-empty.
- Every entry in Response.Events is valid JSON.
- At least one event has `"type":"message_update"`.
- An event with `"type":"agent_end"` exists (confirms stream completed cleanly).

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

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
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
	hasMessageUpdate := false
	hasAgentEnd := false
	for _, raw := range resp.Events {
		if !json.Valid(raw) {
			t.Fatalf("invalid JSON event: %s", string(raw))
		}
		var ev map[string]any
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		typ, _ := ev["type"].(string)
		if typ == "message_update" {
			hasMessageUpdate = true
		}
		if typ == "agent_end" {
			hasAgentEnd = true
		}
	}
	if !hasMessageUpdate {
		t.Fatal("expected at least one event with type 'message_update'")
	}
	if !hasAgentEnd {
		t.Fatal("expected at least one event with type 'agent_end'")
	}
}
```
