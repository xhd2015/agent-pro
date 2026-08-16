---
label: unit
explanation: session list empty map yields empty stdout exit 0
---

## Expected

- Exit code 0.
- Stdout empty (no header when map empty).
- Stderr empty.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout for empty map, got:\n%s", resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
}
```
