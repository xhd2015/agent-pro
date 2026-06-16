## Expected
- Response.Answer references the prior conversation, containing "french" or "capital" (case-insensitive).
- Response.SessionID is non-empty (session persisted across turns).
- Response.Events is non-empty.
- Every entry in Response.Events is valid JSON.
- At least one event has `"type":"text"`.
- At least one event has `"type":"end"` with a non-empty `sessionId`.

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
	if !strings.Contains(lower, "french") && !strings.Contains(lower, "capital") {
		t.Fatalf("expected response to reference prior question, got: %s", resp.Answer)
	}
	if len(resp.Events) == 0 {
		t.Fatal("expected non-empty Events after Ask()")
	}
	hasText := false
	hasEnd := false
	for _, raw := range resp.Events {
		if !json.Valid(raw) {
			t.Fatalf("invalid JSON event: %s", string(raw))
		}
		var ev map[string]any
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		typ, _ := ev["type"].(string)
		if typ == "text" {
			hasText = true
		}
		if typ == "end" {
			sid, _ := ev["sessionId"].(string)
			if sid != "" {
				hasEnd = true
			}
		}
	}
	if !hasText {
		t.Fatal("expected at least one event with type 'text'")
	}
	if !hasEnd {
		t.Fatal("expected at least one event with type 'end' containing sessionId")
	}
}
```
