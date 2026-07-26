## Expected
- Round-trip produces exactly 1 event.
- Event type is `"think"` (`ActionThink`).
- Event text is `"Let me think..."`.
- Event ID is non-empty (FromCrush prepends `"crush:"`).

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
		t.Fatalf("roundtrip think failed: %v", err)
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
	if events[0].Type != types.ActionThink {
		t.Fatalf("expected type %q, got %q", types.ActionThink, events[0].Type)
	}
	if events[0].Text != "Let me think..." {
		t.Fatalf("expected text %q, got %q", "Let me think...", events[0].Text)
	}
	if events[0].ID == "" {
		t.Fatal("expected non-empty ID after round-trip")
	}
}
```
