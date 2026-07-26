## Expected
- Exactly 1 AgentEvent is emitted.
- Its type is `error`.
- Its text is "boom".

```go
import (
	"encoding/json"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Lines) != 1 {
		t.Fatalf("expected 1 AgentEvent (error), got %d lines: %v", len(resp.Lines), resp.Lines)
	}
	var ev eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(resp.Lines[0]), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != eventtypes.ActionError {
		t.Fatalf("expected type %q, got %q", eventtypes.ActionError, ev.Type)
	}
	if ev.Text != "boom" {
		t.Fatalf("expected error text %q, got %q", "boom", ev.Text)
	}
}
```
