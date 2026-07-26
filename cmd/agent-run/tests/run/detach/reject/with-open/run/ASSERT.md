---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Error indicates `--detach` and `--open` cannot be used together (not merely
  unrecognized flag).
- Product-style message:
  `--detach and --open are mutually exclusive; cannot use both`

## Errors

- Mutual exclusion between `--detach` and `--open`.

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
		t.Fatalf("expected non-zero exit for --detach + --open; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, errText, "--detach")

	hasDetach := strings.Contains(errText, "detach")
	hasOpen := strings.Contains(errText, "open")
	hasExclusive := strings.Contains(errText, "mutual") ||
		strings.Contains(errText, "exclusive") ||
		strings.Contains(errText, "cannot") ||
		strings.Contains(errText, "conflict") ||
		strings.Contains(errText, "together") ||
		strings.Contains(errText, "incompatible")
	if !(hasDetach && hasOpen && hasExclusive) {
		t.Fatalf("stderr/stdout should explain --detach vs --open exclusivity:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
