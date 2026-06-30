## Expected

- Exit code 1.
- Stderr indicates invalid session reference format.

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected session-ref validation error, not flag parse failure:\n%s", resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatalf("expected non-empty stderr for malformed session ref")
	}
}
```