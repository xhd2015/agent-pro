## Expected
- One AgentEvent with type `tool_call` and tool `read`.

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	var events []types.AgentEvent
	if err := json.Unmarshal([]byte(resp.Output), &events); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 AgentEvent, got %d: %s", len(events), resp.Output)
	}
	if events[0].Type != types.ActionToolCall {
		t.Fatalf("type = %q, want tool_call", events[0].Type)
	}
	if events[0].Tool != "read" {
		t.Fatalf("tool = %q, want read", events[0].Tool)
	}
}
```