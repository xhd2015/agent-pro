## Expected

- No error.
- Five sessions with Kind tokens fork, sub-f, sub+, sub, main (newest-first).
- Header includes KIND after SESSION ID.
- Each data row for a session ID contains that session's Kind token as a field
  (not merely as a title substring).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertIDsAndKinds(t, resp.Sessions,
		idFork, "fork",
		idSubFork, "sub-f",
		idSubRes, "sub+",
		idSub, "sub",
		idMain, "main",
	)
	assertHeaderKINDColumn(t, resp.Output)

	// For each session, find the data line starting with its ID and require the Kind token.
	lines := strings.Split(resp.Output, "\n")
	for _, s := range resp.Sessions {
		found := false
		for _, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), s.ID) {
				continue
			}
			found = true
			wantKind := sessionKind(s)
			// Token must appear as a whitespace-bounded field on the row.
			fields := strings.Fields(line)
			hasKind := false
			for _, f := range fields {
				if f == wantKind {
					hasKind = true
					break
				}
			}
			if !hasKind {
				t.Fatalf("row for %s missing KIND field %q: %q", s.ID, wantKind, line)
			}
			break
		}
		if !found {
			t.Fatalf("no table row for session %s in output:\n%s", s.ID, resp.Output)
		}
	}
}
```
