## Expected
- Wire contains `agent_thought_chunk`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "agent_thought_chunk")
	assertContains(t, resp.Output, "planning ls output")
}
```
