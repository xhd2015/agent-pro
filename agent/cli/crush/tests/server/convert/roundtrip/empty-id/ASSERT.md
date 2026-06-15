## Expected
- Round-trip produces exactly 1 event.
- Event type is `"message"` (`ActionMessage`).
- Event text is `"test"`.
- Event ID is `"crush:evt_1"` (synthetic ID assigned by `ToCrush`, then prefixed by `FromCrush`).

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip empty-id failed: %v", err)
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
	if events[0].Type != types.ActionMessage {
		t.Fatalf("expected type %q, got %q", types.ActionMessage, events[0].Type)
	}
	if events[0].Text != "test" {
		t.Fatalf("expected text %q, got %q", "test", events[0].Text)
	}
	if events[0].ID != "crush:evt_1" {
		t.Fatalf("expected ID %q, got %q", "crush:evt_1", events[0].ID)
	}
}
```
