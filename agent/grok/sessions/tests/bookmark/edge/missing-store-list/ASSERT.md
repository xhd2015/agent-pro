## Expected

- No error.
- Views empty (len 0).
- Output is empty-style: contains `No bookmarks` or is empty / header-only
  without session id rows (accept either empty message or header with zero data
  rows — must not error).

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
	if len(resp.Views) != 0 {
		t.Fatalf("Views len=%d want 0: %+v", len(resp.Views), resp.Views)
	}
	out := resp.Output
	// Accept "No bookmarks" message (preferred) or empty string / whitespace.
	if strings.TrimSpace(out) == "" {
		return
	}
	if strings.Contains(strings.ToLower(out), "no bookmark") {
		return
	}
	// Header-only tables without data rows are also OK if no session ids appear.
	if strings.Contains(out, fixtureBookmarkSessionID) || strings.Contains(out, fixtureBookmarkSessionID2) {
		t.Fatalf("empty list output unexpectedly contains session ids:\n%s", out)
	}
}
```
