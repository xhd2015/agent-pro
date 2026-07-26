---
label: e2e
---

## Expected
- Exit code 0.
- Stdout contains both "before" and "after" messages.
- Exactly 2 JSON lines in stdout (sleep produces no output).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    lines := parseJSONLines(t, resp.Stdout)
    if len(lines) != 2 {
        t.Fatalf("expected exactly 2 output lines (sleep produces none), got %d lines:\n%s", len(lines), resp.Stdout)
    }
    assertContains(t, resp.Stdout, `"text":"before"`)
    assertContains(t, resp.Stdout, `"text":"after"`)
}
```
