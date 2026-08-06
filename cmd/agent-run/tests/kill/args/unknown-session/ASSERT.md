## Expected

- Exit code non-zero (typically 1).
- `kill` is a recognized command (not `unknown command: kill`).
- Stderr indicates session not found or expired.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for unknown session, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	// Stay RED until kill is implemented.
	if strings.Contains(lower, "unknown command") {
		t.Fatalf("kill not implemented yet (unknown command); want session not found / expired, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(lower, "not found") && !strings.Contains(lower, "expired") {
		t.Fatalf("expected stderr to mention not found / expired, got:\n%s", resp.Stderr)
	}
}
```
