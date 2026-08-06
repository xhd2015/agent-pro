## Expected

- No error.
- Exactly one view; Orphaned=true.
- Title remains stored snapshot.
- Warnings mention session id / not found.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Views) != 1 {
		t.Fatalf("Views len=%d want 1: %+v", len(resp.Views), resp.Views)
	}
	v := resp.Views[0]
	assertEqualString(t, "SessionID", v.SessionID, req.SessionID)
	assertEqualBool(t, "Orphaned", v.Orphaned, true)
	assertEqualString(t, "Title", v.Title, req.StoredTitle)
	if !warningHasSession(resp.Warnings, req.SessionID) {
		ok := false
		for _, w := range resp.Warnings {
			if strings.Contains(w, req.SessionID) {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("expected orphan warning mentioning %s, got %v", req.SessionID, resp.Warnings)
		}
	}
}
```
