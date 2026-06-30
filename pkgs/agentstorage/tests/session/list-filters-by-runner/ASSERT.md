## Expected

- `ListSessions` returns exactly one session.
- The returned session belongs to `fake-opencode` with id `sess_opencode`.
- No `fake-codex` session appears in the list.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d: %+v", len(resp.Sessions), resp.Sessions)
	}
	assertEqual(t, "Runner", resp.Sessions[0].Runner, req.Runner)
	assertEqual(t, "SessionID", resp.Sessions[0].SessionID, req.SessionID)
}
```