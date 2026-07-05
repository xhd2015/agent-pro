## Expected
- Wire `tool_call` line includes a non-empty generated `toolCallId`.

```go
import (
	"testing"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "tool_call")
	if !strings.Contains(resp.Output, `"toolCallId":"`) {
		t.Fatalf("expected generated toolCallId in wire:\n%s", resp.Output)
	}
	assertNotContains(t, resp.Output, `"toolCallId":""`)
}
```
