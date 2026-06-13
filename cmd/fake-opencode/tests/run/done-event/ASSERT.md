## Expected
- The command succeeds.
- The done event is emitted with session ID and `"done":true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"done":true`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_done"`)
}
```
