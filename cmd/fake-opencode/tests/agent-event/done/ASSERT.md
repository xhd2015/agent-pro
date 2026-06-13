## Expected
- The command succeeds.
- stdout contains an opencode event with `"done":true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"sessionID":"sess_done"`)
    assertContains(t, resp.Stdout, `"done":true`)
}
```
