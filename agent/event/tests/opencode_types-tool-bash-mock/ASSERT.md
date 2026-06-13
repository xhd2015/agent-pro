## Expected
- One opencode event: type `tool_use` with session ID, tool name `bash`, and state containing output and exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"tool_use"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_001"`)
	assertContains(t, resp.Stdout, `"tool":"bash"`)
	assertContains(t, resp.Stdout, `"output":"hello"`)
	assertContains(t, resp.Stdout, `"exit_code":0`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
