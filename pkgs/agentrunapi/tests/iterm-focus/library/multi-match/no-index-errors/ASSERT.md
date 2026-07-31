## Expected

- `FocusSession` returns a non-nil error (ambiguous / multi / need --index).
- `FocusITerm` is never called.
- `FindITermForSession` path exposes at least 2 candidates (via `resp.Candidates`).

## Errors

- Non-nil from `FocusSession`.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertError(t, err)
	assertNoFocus(t, resp)
	if len(resp.Candidates) < 2 {
		t.Fatalf("expected >=2 candidates for multi, got %d: %+v", len(resp.Candidates), resp.Candidates)
	}
	msg := strings.ToLower(err.Error())
	// Prefer error that steers user to --index; soft if wording differs.
	if !strings.Contains(msg, "index") && !strings.Contains(msg, "multiple") &&
		!strings.Contains(msg, "multi") && !strings.Contains(msg, "ambiguous") {
		t.Logf("multi error (acceptable): %v", err)
	}
}
```
