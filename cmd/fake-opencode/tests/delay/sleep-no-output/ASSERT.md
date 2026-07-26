---
label: e2e
---

## Expected
- Exit code 0.
- Exactly 1 JSON line in stdout (the message, sleep yields nothing).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    lines := parseJSONLines(t, resp.Stdout)
    if len(lines) != 1 {
        t.Fatalf("expected exactly 1 output line (sleep produces none), got %d lines:\n%s", len(lines), resp.Stdout)
    }
    assertContains(t, resp.Stdout, `"text":"only"`)
}
```
