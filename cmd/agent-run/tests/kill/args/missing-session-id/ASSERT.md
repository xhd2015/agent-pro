## Expected

- Exit code non-zero (typically 1).
- `kill` is a recognized command (not `unknown command: kill`).
- Stderr indicates missing session-id / usage for `kill`.

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
		t.Fatalf("expected non-zero exit for missing session-id, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	// Stay RED until kill is implemented: "unknown command: kill" is not the contract.
	if strings.Contains(lower, "unknown command") {
		t.Fatalf("kill not implemented yet (unknown command); want missing session-id validation, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "usage") && !strings.Contains(lower, "require") {
		t.Fatalf("expected error to mention session-id / usage / require, got:\n%s", resp.Stderr)
	}
}
```
