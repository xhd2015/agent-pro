## Expected
- One opencode event: type `reasoning` with session ID and think text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"reasoning"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_001"`)
	assertContains(t, resp.Stdout, `"thinking about the problem"`)
	assertContains(t, resp.Stdout, `"id":"evt_1"`)
}
```
