## Expected

- No error.
- Exactly one view; Orphaned=false.
- Title and NumChatMessages match **live** session (Find recovery).
- SessionDir non-empty (recovered absolute path).
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
	assertEqualBool(t, "Orphaned", v.Orphaned, false)
	assertEqualString(t, "Title", v.Title, req.LiveTitle)
	assertEqualInt(t, "NumChatMessages", v.NumChatMessages, req.LiveNumChatMessages)
	if v.SessionDir == "" {
		t.Fatal("SessionDir empty after heavy Find recovery")
	}
	if v.Title == req.StoredTitle {
		t.Fatalf("Title still stored snapshot %q; heavy should recover live", req.StoredTitle)
	}
	if len(resp.Warnings) != 0 {
		t.Fatalf("Warnings=%v want empty after recover", resp.Warnings)
	}
}
```
