## Expected
- One opencode event: type `text` with session ID and message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"here is the result"`)
	assertContains(t, resp.Output, `"id":"evt_2"`)
}
```
