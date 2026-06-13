## Expected
- The command succeeds.
- stdout contains an opencode text event with session ID and the message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"text"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_msg"`)
    assertContains(t, resp.Stdout, `"all tasks complete"`)
}
```
