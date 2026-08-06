## Expected

- No error.
- Exactly one view; Orphaned=true.
- Title still the stored snapshot (`orphan snapshot title`).
- Warnings include a message mentioning the session id and not found
  (under GROK_HOME).

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
	assertEqualBool(t, "Orphaned", v.Orphaned, true)
	assertEqualString(t, "Title", v.Title, req.Title)
	assertEqualString(t, "SessionID", v.SessionID, req.SessionID)
	if !warningHasSession(resp.Warnings, req.SessionID) {
		// also accept warning text without helper match if phrasing differs slightly
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
