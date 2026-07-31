## Expected

- `FocusSession` returns a non-nil error (no iTerm candidate / none found).
- `FocusITerm` is never called.
- Optional: error text is informative (contains "not found" / "no" / "none" case-insensitively) — soft check.

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
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "not found") &&
		!strings.Contains(msg, "no ") &&
		!strings.Contains(msg, "none") &&
		!strings.Contains(msg, "no match") &&
		!strings.Contains(msg, "candidate") {
		// Soft: still pass on any error; log shape for implementer guidance.
		t.Logf("none-found error (acceptable any non-nil): %v", err)
	}
}
```
