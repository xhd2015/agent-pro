## Expected
- Output JSON array contains three events in order: think, message, tool_call.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"thinking..."`)
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"text":"result"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
}
```
