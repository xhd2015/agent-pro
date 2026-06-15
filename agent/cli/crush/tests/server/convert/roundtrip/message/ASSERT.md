## Expected
- Round-trip produces exactly 1 event.
- Event type is `"message"` (`ActionMessage`).
- Event text is `"Hello world"`.
- Event ID is non-empty and starts with `"crush:"`.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip message failed: %v", err)
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
	if events[0].Text != "Hello world" {
		t.Fatalf("expected text %q, got %q", "Hello world", events[0].Text)
	}
	if !strings.HasPrefix(events[0].ID, "crush:") {
		t.Fatalf("expected ID to start with 'crush:', got %q", events[0].ID)
	}
}
```
