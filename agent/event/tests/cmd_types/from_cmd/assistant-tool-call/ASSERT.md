## Expected
- Output JSON array contains one event with `"type":"tool_call"`, `"tool":"bash"` and `"tool_input":{"command":"echo hi"}`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"tool_input":{"command":"echo hi"}`)
}
```
