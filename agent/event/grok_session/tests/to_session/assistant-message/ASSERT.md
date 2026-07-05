## Expected
- Wire contains `agent_message_chunk` with assistant text.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "agent_message_chunk")
	assertContains(t, resp.Output, "Here is the answer")
}
```
