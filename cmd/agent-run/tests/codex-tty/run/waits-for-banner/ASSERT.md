## Expected

- Exit code 0.
- Stderr does not contain `banner not detected` (or equivalent failure).
- Captured output contains `hi` — prompt was injected after banner appeared.

## Errors

- Must not fail with codex TUI banner timeout.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "banner not detected") || strings.Contains(lower, "tui banner") {
		t.Fatalf("run failed banner wait, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "hi") {
		t.Fatalf("expected prompt injected after banner; stdout:\n%s", resp.Stdout)
	}
}
```