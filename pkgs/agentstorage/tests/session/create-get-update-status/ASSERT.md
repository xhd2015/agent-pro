## Expected

- `GetSession` after update returns status `finished`.
- Session meta preserves runner, session_id, and model.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.Session == nil {
		t.Fatal("expected session, got nil")
	}
	assertEqual(t, "Status", resp.Session.Meta.Status, "finished")
	assertEqual(t, "Runner", resp.Session.Meta.Runner, req.Runner)
	assertEqual(t, "SessionID", resp.Session.Meta.SessionID, req.SessionID)
	assertEqual(t, "Model", resp.Session.Meta.Model, req.Model)
}
```