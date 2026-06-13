## Expected
- One opencode event: type `done` with session ID and `"done":true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"done":true`)
}
```
