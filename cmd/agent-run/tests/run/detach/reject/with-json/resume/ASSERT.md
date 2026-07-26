---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Error indicates `--detach` and `--json` mutually exclusive (not unknown flag).

## Errors

- Mutual exclusion between `--detach` and `--json` on resume.

## Exit Code

non-zero

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for resume --detach --json; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, errText, "--detach")
	hasDetach := strings.Contains(errText, "detach")
	hasJSON := strings.Contains(errText, "json")
	hasExclusive := strings.Contains(errText, "mutual") ||
		strings.Contains(errText, "exclusive") ||
		strings.Contains(errText, "cannot") ||
		strings.Contains(errText, "conflict") ||
		strings.Contains(errText, "together")
	if !(hasDetach && hasJSON && hasExclusive) {
		t.Fatalf("stderr/stdout should explain resume --detach vs --json exclusivity:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
