## Expected
- Output JSON array contains one event with `"type":"tool_call"`, `"tool":"Bash"`, and `"tool_input":{"command":"ls"}`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"Bash"`)
	assertContains(t, resp.Output, `"tool_input":{"command":"ls"}`)
}
```
