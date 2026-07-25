## Expected

- Exit code 0.
- Combined stdout+stderr mentions **`proc`** (top-level command).
- Combined text mentions **`resolve`** (resolve agent session from process id).
- Stdout or combined help ends with a trailing newline on the primary stream
  when present (soft: do not fail solely on newline if stderr-only).

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0

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
	assertExitCode(t, resp, 0)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "proc")
	assertContainsFold(t, combined, "resolve")
	// Prefer stdout as the help channel when non-empty.
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("root help stdout should end with newline; got %q", resp.Stdout)
	}
}
```
