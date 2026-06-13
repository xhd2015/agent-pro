## Expected
- JSON contains `"type":"tool_use"`, `"sessionID":"sess_t1"`, and a `part` with tool name, state, output, exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"tool_use"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_t1"`)
	assertContains(t, resp.Stdout, `"tool":"bash"`)
	assertContains(t, resp.Stdout, `"output":"hello"`)
	assertContains(t, resp.Stdout, `"exit_code":0`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
