## Expected

- `TailUpdatesFromOffset` emits two `ActionToolCall` events.
- Both events carry `tool_call_id=call_read_1`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	tools := toolEvents(resp.Events)
	if len(tools) != 2 {
		t.Fatalf("tool event count: got %d want 2\n%s", len(tools), formatEvents(resp.Events))
	}
	for i, ev := range tools {
		if ev.ToolCallID != "call_read_1" {
			t.Fatalf("tools[%d] tool_call_id: got %q want call_read_1 (%#v)", i, ev.ToolCallID, ev)
		}
		if ev.Type != types.ActionToolCall {
			t.Fatalf("tools[%d] type: got %s want tool_call", i, ev.Type)
		}
	}
}
```