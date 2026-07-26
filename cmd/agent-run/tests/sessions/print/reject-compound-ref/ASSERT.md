---
label: e2e
---

## Expected

- Exit code 1.
- Stderr indicates invalid session reference (slash / compound form not allowed).
- Not a flag parse failure.

## Exit Code

1

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected session-ref validation error, not flag parse failure:\n%s", resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatalf("expected non-empty stderr for compound session ref")
	}
	lower := strings.ToLower(resp.Stderr)
	// should mention invalid / reference / session
	if !strings.Contains(lower, "invalid") &&
		!strings.Contains(lower, "reference") &&
		!strings.Contains(lower, "session") &&
		!strings.Contains(lower, "/") {
		t.Fatalf("expected invalid-ref message, got:\n%s", resp.Stderr)
	}
}
```
