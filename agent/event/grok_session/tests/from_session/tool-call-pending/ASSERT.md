## Expected
- One `ActionToolCall` with `tool_call_id=call_read_1`, `status=pending`, `turn_index=0`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertEventCount(t, resp.Events, 1)
	ev := resp.Events[0]
	if ev.Type != types.ActionToolCall {
		t.Fatalf("type: got %s want tool_call", ev.Type)
	}
	if ev.ToolCallID != "call_read_1" {
		t.Fatalf("tool_call_id: got %q want call_read_1", ev.ToolCallID)
	}
	if grokStatus(ev) != "pending" {
		t.Fatalf("status: got %q want pending", grokStatus(ev))
	}
	if ev.Text != "README.md" {
		t.Fatalf("title text: got %q", ev.Text)
	}
}
```
