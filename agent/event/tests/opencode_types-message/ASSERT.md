## Expected
- One opencode event: type `text` with session ID and message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"text"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_001"`)
	assertContains(t, resp.Stdout, `"here is the result"`)
	assertContains(t, resp.Stdout, `"id":"evt_2"`)
}
```
