## Expected

- No error.
- View non-nil; SessionID and AgentRunner=grok match fixture.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.View == nil {
		t.Fatal("View is nil")
	}
	assertEqualString(t, "SessionID", resp.View.SessionID, req.SessionID)
	assertEqualString(t, "AgentRunner", resp.View.AgentRunner, "grok")
	assertEqualString(t, "Title", resp.View.Title, req.Title)
}
```
