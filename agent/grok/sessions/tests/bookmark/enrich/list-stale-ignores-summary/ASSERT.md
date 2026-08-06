## Expected

- No error.
- Exactly one view; Orphaned=false (not computed / always false for off).
- Title and NumChatMessages match **stored** catalog snapshot (ignore live summary).
- Warnings empty (no orphan computation).

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
	assertEqualString(t, "Title", v.Title, req.StoredTitle)
	assertEqualInt(t, "NumChatMessages", v.NumChatMessages, req.StoredNumChatMessages)
	assertEqualBool(t, "Orphaned", v.Orphaned, false)
	if v.Title == req.LiveTitle {
		t.Fatalf("Title picked up live summary %q; EnrichOff must keep store", req.LiveTitle)
	}
	if len(resp.Warnings) != 0 {
		t.Fatalf("Warnings=%v want empty under EnrichOff", resp.Warnings)
	}
}
```
