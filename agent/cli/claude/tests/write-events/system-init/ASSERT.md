## Expected
- Exactly 1 AgentEvent is emitted.
- Its type is `step_start`.

```go
import (
	"encoding/json"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Lines) != 1 {
		t.Fatalf("expected 1 AgentEvent (step_start), got %d lines: %v", len(resp.Lines), resp.Lines)
	}
	var ev eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(resp.Lines[0]), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != eventtypes.ActionStepStart {
		t.Fatalf("expected type %q, got %q", eventtypes.ActionStepStart, ev.Type)
	}
}
```
