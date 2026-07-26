---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions the unknown/invalid kind (flexible wording: kind, unknown, invalid).

## Errors

- Invalid `--kind` value is a user error, not a process-list failure.

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	ok := strings.Contains(combined, "kind") ||
		strings.Contains(combined, "unknown") ||
		strings.Contains(combined, "invalid") ||
		strings.Contains(combined, "not-a-real-kind")
	if !ok {
		t.Fatalf("expected unknown-kind error on stderr/stdout; stderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
