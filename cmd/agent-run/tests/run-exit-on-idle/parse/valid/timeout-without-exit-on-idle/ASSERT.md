## Expected

- `ParseRunIdle(false, "2s")` succeeds (timeout ignored when the flag is off).
- Exit code is 0.
- Stderr is empty (no `Error:`).
- `Enabled` is false.

## Side Effects

- None (parse-only; no TTY).

## Errors

- None.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("timeout without --exit-on-idle must parse OK; exit %d stderr %q", resp.ExitCode, resp.Stderr)
	}
	if resp.ErrString != "" {
		t.Fatalf("unexpected parse error: %s", resp.ErrString)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr must be empty on parse OK; got %q", resp.Stderr)
	}
	if resp.Enabled {
		t.Fatal("ParseRunIdle(false, 2s) must return enabled=false")
	}
}
```
