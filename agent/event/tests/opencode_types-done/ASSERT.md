## Expected
- One opencode event: type `done` with session ID and `"done":true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"done"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_001"`)
	assertContains(t, resp.Stdout, `"done":true`)
}
```
