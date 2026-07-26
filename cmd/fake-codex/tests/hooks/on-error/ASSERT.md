---
label: e2e
---

## Expected
- The process exits with the configured nonzero code.
- The `on_error` hook fires.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode != 9 {
        t.Fatalf("exit code = %d, want 9; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.HookLog, `"event":"OnError"`)
}
```

