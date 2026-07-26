---
label: e2e
---

## Expected
- The command exits with code 7.
- stdout and stderr preserve configured content.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode != 7 {
        t.Fatalf("exit code = %d, want 7; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"before failure"`)
    assertContains(t, resp.Stderr, "planned failure")
}
```

