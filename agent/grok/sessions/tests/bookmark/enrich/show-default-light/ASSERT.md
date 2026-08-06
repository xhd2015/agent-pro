## Expected

- No error; View non-nil; Orphaned=false.
- Title and NumChatMessages match live mutated summary.
- SessionID and SessionDir set.

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
	v := resp.View
	assertEqualString(t, "SessionID", v.SessionID, req.SessionID)
	assertEqualString(t, "AgentRunner", v.AgentRunner, "grok")
	assertEqualBool(t, "Orphaned", v.Orphaned, false)
	assertEqualString(t, "Title", v.Title, req.LiveTitle)
	assertEqualInt(t, "NumChatMessages", v.NumChatMessages, req.LiveNumChatMessages)
	if v.SessionDir == "" {
		t.Fatal("SessionDir empty after show light refresh")
	}
	if v.Title == req.StoredTitle {
		t.Fatalf("Title still stored snapshot %q; show light should refresh", req.StoredTitle)
	}
}
```
