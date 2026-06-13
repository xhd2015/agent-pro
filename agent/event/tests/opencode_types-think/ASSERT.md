## Expected
- One opencode event: type `reasoning` with session ID and think text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"thinking about the problem"`)
	assertContains(t, resp.Output, `"id":"evt_1"`)
}
```
