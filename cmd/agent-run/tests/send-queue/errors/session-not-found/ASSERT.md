---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions session not found.
- Stdout does not contain a message id line.

## Exit Code

1

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertExitCode(t, resp, 1)
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected no stdout id line, got %q", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "not found") {
		t.Fatalf("stderr should mention not found, got: %s", resp.Stderr)
	}
}
```