---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Error indicates `--no-submit` requires `--open` (detach does not satisfy it).
- Not merely unrecognized flag for `--detach` or `--no-submit`.

## Errors

- `--no-submit` without `--open` even when `--detach` is set.

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
		t.Fatalf("expected non-zero exit for --detach --no-submit without --open; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, errText, "--detach")
	assertNotUnknownFlag(t, errText, "--no-submit")

	hasNoSubmit := strings.Contains(errText, "--no-submit") ||
		strings.Contains(errText, "no-submit") ||
		(strings.Contains(errText, "no") && strings.Contains(errText, "submit"))
	hasOpen := strings.Contains(errText, "--open") || strings.Contains(errText, "open")
	hasRequires := strings.Contains(errText, "require") ||
		strings.Contains(errText, "must") ||
		strings.Contains(errText, "need") ||
		strings.Contains(errText, "only with")
	if !(hasNoSubmit && hasOpen && hasRequires) {
		t.Fatalf("stderr/stdout should explain --no-submit requires --open (detach not enough):\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
