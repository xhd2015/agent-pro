## Expected
- One opencode event: type `error` with session ID and error message in nested structure.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"something went wrong"`)
	assertContains(t, resp.Output, `"name":"Error"`)
}
```
