## Expected
- Wire has both `tool_call` and `tool_call_update` with same id.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "tool_call")
	assertWireHasSessionUpdate(t, resp.WireLines, "tool_call_update")
	assertWireToolCallID(t, resp.WireLines, "call_read_1")
	assertContains(t, resp.Output, "package main")
}
```
