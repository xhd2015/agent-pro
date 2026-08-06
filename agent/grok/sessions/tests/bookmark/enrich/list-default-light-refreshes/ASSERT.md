## Expected

- No error.
- Exactly one view; Orphaned=false.
- Title and NumChatMessages match **live** mutated summary (not store snapshot).
- SessionDir still the stored/live dir.
- Warnings empty.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Views) != 1 {
		t.Fatalf("Views len=%d want 1: %+v", len(resp.Views), resp.Views)
	}
	v := resp.Views[0]
	assertEqualString(t, "SessionID", v.SessionID, req.SessionID)
	assertEqualString(t, "AgentRunner", v.AgentRunner, "grok")
	assertEqualString(t, "Title", v.Title, req.LiveTitle)
	assertEqualInt(t, "NumChatMessages", v.NumChatMessages, req.LiveNumChatMessages)
	assertEqualBool(t, "Orphaned", v.Orphaned, false)
	if v.SessionDir == "" {
		t.Fatal("SessionDir empty after light refresh")
	}
	if len(resp.Warnings) != 0 {
		t.Fatalf("Warnings=%v want empty", resp.Warnings)
	}
	// Must not keep stale store snapshot.
	if v.Title == req.StoredTitle {
		t.Fatalf("Title still stored snapshot %q; light should refresh from summary", req.StoredTitle)
	}
}
```
