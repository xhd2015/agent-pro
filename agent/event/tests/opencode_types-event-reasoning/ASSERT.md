## Expected
- JSON contains `"type":"reasoning"`, `"sessionID":"sess_r1"`, and a `part` with reasoning text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"reasoning"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_r1"`)
	assertContains(t, resp.Stdout, `"id":"evt_r1"`)
	assertContains(t, resp.Stdout, `"thinking step by step"`)
}
```
