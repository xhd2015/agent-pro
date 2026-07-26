## Expected
- Round-trip produces exactly 1 event.
- Event type is `"error"` (`ActionError`).
- Event text is `"something failed"`.
- Event ID is empty (no run ID in source event).

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip error failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var events []types.AgentEvent
	if err := json.Unmarshal([]byte(resp.Output), &events); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, resp.Output)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != types.ActionError {
		t.Fatalf("expected type %q, got %q", types.ActionError, events[0].Type)
	}
	if events[0].Text != "something failed" {
		t.Fatalf("expected text %q, got %q", "something failed", events[0].Text)
	}
}
```
