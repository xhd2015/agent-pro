## Expected

- Exit code non-zero (typically 1).
- `takeover` is a recognized command (not `unknown command: takeover`).
- Stderr indicates missing session-id / usage for `takeover`.

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
	assertTakeoverRecognized(t, resp.Stderr)
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "usage") && !strings.Contains(lower, "require") {
		t.Fatalf("expected error to mention session-id / usage / require, got:\n%s", resp.Stderr)
	}
}
```
