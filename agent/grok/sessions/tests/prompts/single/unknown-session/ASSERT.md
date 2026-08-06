## Errors

- Error containing `grok session not found` and the requested session id.
- Single is nil (or ignored when Err set).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "grok session not found", req.SessionID)
}
```
