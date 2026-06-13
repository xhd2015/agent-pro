## Expected
- JSON contains `"type":"error"`, `"sessionID":"sess_e1"`, and a nested `error` object with name and message.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"error"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_e1"`)
	assertContains(t, resp.Stdout, `"name":"Error"`)
	assertContains(t, resp.Stdout, `"something went wrong"`)
}
```
