## Expected

- `CreateSession` stores `workspace` on session meta.
- `GetSession` returns the same `workspace` value.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.Session == nil {
		t.Fatal("expected session, got nil")
	}
	assertEqual(t, "Workspace", resp.Session.Meta.Workspace, req.Workspace)
	assertEqual(t, "SessionID", resp.Session.Meta.SessionID, req.SessionID)
}
```