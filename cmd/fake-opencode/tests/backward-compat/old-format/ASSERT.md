## Expected
- The command succeeds.
- The legacy opencode event is emitted exactly as configured.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"text"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_old"`)
    assertContains(t, resp.Stdout, `"legacy opencode ok"`)
}
```
