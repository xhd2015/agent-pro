---
label: e2e
---

## Expected
- The command exits with code 4.
- stdout contains an opencode error event with nested `error` structure (not raw text pass-through).
- stderr contains the scripted error string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode != 4 {
        t.Fatalf("exit code = %d, want 4; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"type":"error"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_err"`)
    assertContains(t, resp.Stdout, `"operation failed"`)
    assertContains(t, resp.Stdout, `"error":`)
    assertContains(t, resp.Stdout, `"message":`)
    assertContains(t, resp.Stderr, "planned failure")
}
```
