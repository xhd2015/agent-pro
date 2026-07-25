## Expected

- Exit code is non-zero.
- Stderr is non-empty and indicates failure (contains `not found` and/or `pid`,
  case-insensitive).
- Prefer stdout empty or free of a hard session success claim; not strictly
  locked beyond exit + stderr.

## Side Effects

- None.

## Errors

- Pid-not-found class error on stderr.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatalf("expected error on stderr, got empty\nstdout:\n%s", resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "not found") && !strings.Contains(low, "pid") {
		t.Fatalf("stderr should mention not found and/or pid:\n%s", resp.Stderr)
	}
}
```
