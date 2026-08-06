## Expected

- Error containing `grok session not found` and the requested session id.
- `Status` is nil.

## Errors

- `grok session not found: 019f283a-eeee-7eee-eeee-eeeeeeeeee99`

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "grok session not found", req.SessionID)
	if resp.Status != nil {
		t.Fatalf("expected nil Status on error, got %+v", resp.Status)
	}
}
```
