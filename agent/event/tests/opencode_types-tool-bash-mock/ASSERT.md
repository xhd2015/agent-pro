## Expected
- One opencode event: type `tool_use` with session ID, tool name `bash`, and state containing output and exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_use"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"output":"hello"`)
	assertContains(t, resp.Output, `"exit_code":0`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
