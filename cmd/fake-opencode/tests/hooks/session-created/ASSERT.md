## Expected
- The `session.created` hook fires.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"session.created"`)
    assertContains(t, resp.HookLog, `"session_id":"sess_hook"`)
}
```

