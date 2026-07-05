## Expected
- Wire contains `tool_call` with matching toolCallId and status pending.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "tool_call")
	assertWireToolCallID(t, resp.WireLines, "call_read_1")
	assertContains(t, resp.Output, `"status":"pending"`)
}
```
