## Expected

- `Backup` returns an error.
- Error contains `grok session not found` and the session id.
- Result is nil.
- OutDir not created / no payload.

## Errors

- `grok session not found: 019f283a-eeee-7eee-eeee-eeeeeeeeee99`

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "grok session not found", req.SessionID)
	if resp.Result != nil {
		t.Fatalf("expected nil Result, got %+v", resp.Result)
	}
	assertNoPayloadUnder(t, req.OutDir)
	assertPathMissing(t, req.OutDir)
}
```
