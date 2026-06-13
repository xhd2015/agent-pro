## Expected
- The command succeeds.
- stdout contains an opencode tool_use event for bash with session ID.
- The tool_use part has status completed and contains the real command output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_bash"`)
    assertContains(t, resp.Stdout, `"tool":"bash"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
    assertContains(t, resp.Stdout, `"hello-opencode"`)
}
```
