## Expected

- No error.
- One session returned.
- Output header line contains columns in order: SESSION ID, KIND, LAST ACTIVE
  (TITLE, MSGS, CWD also present).
- Header is not the old layout without KIND.

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
	assertSessionIDs(t, resp.Sessions, idMain)
	if strings.TrimSpace(resp.Output) == "" {
		t.Fatal("expected table output, got empty")
	}
	assertHeaderKINDColumn(t, resp.Output)
	// Ensure title/msgs/cwd headers still present after KIND insertion.
	header := resp.Output
	if i := strings.IndexByte(resp.Output, '\n'); i >= 0 {
		header = resp.Output[:i]
	}
	for _, col := range []string{"TITLE", "MSGS", "CWD"} {
		if !strings.Contains(header, col) {
			t.Fatalf("header missing %q: %q", col, header)
		}
	}
}
```
