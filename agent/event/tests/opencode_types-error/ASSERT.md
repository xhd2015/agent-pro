## Expected
- One opencode event: type `error` with session ID and error message in nested structure.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"error"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_001"`)
	assertContains(t, resp.Stdout, `"something went wrong"`)
	assertContains(t, resp.Stdout, `"name":"Error"`)
}
```
