---
label: e2e
---

## Expected

- Must **not** fail with "prompt is required" — empty followup means resume-only reopen.
- Exit may still be non-zero if the TTY runner cannot start in this fixture (no real grok);
  that is OK as long as the error is not about a missing prompt.

## Exit Code

any (must not be the old "prompt is required" gate)

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(combined, "prompt is required") ||
		strings.Contains(combined, "prompt required") ||
		strings.Contains(combined, "requires a prompt") {
		t.Fatalf("resume without followup must not require a prompt; got:\n%s", combined)
	}
	// If the command succeeded, good. If it failed, ensure it got past the gate
	// (e.g. unknown binary / tty errors are acceptable for this leaf).
	if resp.ExitCode != 0 {
		t.Logf("resume no-prompt exited %d (gate passed); stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}
```
