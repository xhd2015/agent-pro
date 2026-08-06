## Expected

- No error.
- Exactly one view with SessionID = primary fixture id (has both tags).
- Secondary session (backup only) not returned.

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
	assertEqualString(t, "SessionID", resp.Views[0].SessionID, fixtureBookmarkSessionID)
	if _, ok := viewBySession(resp.Views, fixtureBookmarkSessionID2); ok {
		t.Fatal("session with only backup tag should not match AND filter")
	}
}
```
