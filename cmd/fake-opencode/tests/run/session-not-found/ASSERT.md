---
label: e2e
---

## Expected
- The command exits with non-zero exit code.
- Stdout contains an error event with type "error".
- The error event references the missing session.
- Stderr contains "Session not found".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatalf("expected non-zero exit code for missing session\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"type":"error"`)
    assertContains(t, resp.Stdout, `"sessionID":"no-such-session"`)
    assertContains(t, resp.Stderr, "Session not found")
}
```
