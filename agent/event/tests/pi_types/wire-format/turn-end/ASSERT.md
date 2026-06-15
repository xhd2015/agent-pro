## Expected
- Output contains turn_end type, assistant message, and tool results.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"turn_end"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"role":"toolResult"`)
	assertContains(t, resp.Output, `"toolCallId":"call_1"`)
}
```
