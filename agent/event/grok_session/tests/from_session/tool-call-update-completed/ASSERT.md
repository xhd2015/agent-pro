## Expected
- Two tool events sharing `tool_call_id`; second has `Output` and `status=completed`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertEventCount(t, resp.Events, 2)
	if resp.Events[0].ToolCallID != "call_read_1" || grokStatus(resp.Events[0]) != "pending" {
		t.Fatalf("pending event: %#v", resp.Events[0])
	}
	ev := resp.Events[1]
	if ev.ToolCallID != "call_read_1" || grokStatus(ev) != "completed" {
		t.Fatalf("completed event: %#v", ev)
	}
	if ev.Output != "package main" {
		t.Fatalf("output: got %q want package main", ev.Output)
	}
}
```
