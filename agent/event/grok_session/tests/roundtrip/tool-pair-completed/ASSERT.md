## Expected
- tool_call_id, status=completed, and Output preserved semantically.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSemanticEqualEvents(t, resp.Events1, resp.Events2)
	var completed *types.AgentEvent
	for i := range resp.Events1 {
		ev := &resp.Events1[i]
		if ev.Type == types.ActionToolCall && grokStatus(*ev) == "completed" {
			completed = ev
		}
	}
	if completed == nil {
		t.Fatal("missing completed tool event")
	}
	if completed.ToolCallID != "call_read_1" || completed.Output != "package main" {
		t.Fatalf("completed tool: %#v", completed)
	}
}
```
