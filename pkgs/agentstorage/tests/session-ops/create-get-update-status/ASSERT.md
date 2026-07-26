## Expected

- `GetSession` after update returns status `finished`.
- Session meta preserves runner, session_id, and model.
- Durable path is flat: `sessions/<session_id>/meta.json` exists.
- Nested path `sessions/<runner>/<session_id>/` does not exist.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.Session == nil {
		t.Fatal("expected session, got nil")
	}
	assertEqual(t, "Status", resp.Session.Meta.Status, "finished")
	assertEqual(t, "Runner", resp.Session.Meta.Runner, req.Runner)
	assertEqual(t, "SessionID", resp.Session.Meta.SessionID, req.SessionID)
	assertEqual(t, "Model", resp.Session.Meta.Model, req.Model)

	flatMeta := filepath.Join(resp.ResolvedHome, "sessions", req.SessionID, "meta.json")
	if _, err := os.Stat(flatMeta); err != nil {
		t.Fatalf("expected flat meta at %q: %v", flatMeta, err)
	}
	nestedDir := filepath.Join(resp.ResolvedHome, "sessions", req.Runner, req.SessionID)
	if st, err := os.Stat(nestedDir); err == nil && st.IsDir() {
		t.Fatalf("unexpected nested runner session dir %q", nestedDir)
	}
}
```
