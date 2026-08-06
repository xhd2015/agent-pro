## Expected

- Error containing `grok session not found` and the session id.
- Files slice empty / Dir empty (or ignored when Err set).

## Errors

- `grok session not found: 019f283a-eeee-7eee-eeee-eeeeeeeeee88`

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "grok session not found", req.SessionID)
}
```
