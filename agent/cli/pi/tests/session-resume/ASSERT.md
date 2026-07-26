## Expected
- Response.Answer references the prior conversation, containing "french" or "capital".
- Response.SessionID is non-empty (a session was created and reused).
- Response.Events is non-empty.
- Every entry in Response.Events is valid JSON.
- At least one event has `"type":"message_start"`.
- At least one event has `"type":"message_update"`.
- At least one event has `"type":"message_end"`.
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
	if !strings.Contains(lower, "french") && !strings.Contains(lower, "capital") {
		t.Fatalf("expected response to reference prior question, got: %s", resp.Answer)
	}
	if len(resp.Events) == 0 {
		t.Fatal("expected non-empty Events after Ask()")
	}
	hasMessageStart := false
	hasMessageUpdate := false
	hasMessageEnd := false
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
		switch typ {
		case "message_start":
			hasMessageStart = true
		case "message_update":
			hasMessageUpdate = true
		case "message_end":
			hasMessageEnd = true
		case "agent_end":
			hasAgentEnd = true
		}
	}
	if !hasMessageStart {
		t.Fatal("expected at least one event with type 'message_start'")
	}
	if !hasMessageUpdate {
		t.Fatal("expected at least one event with type 'message_update'")
	}
	if !hasMessageEnd {
		t.Fatal("expected at least one event with type 'message_end'")
	}
	if !hasAgentEnd {
		t.Fatal("expected at least one event with type 'agent_end'")
	}
}
```
