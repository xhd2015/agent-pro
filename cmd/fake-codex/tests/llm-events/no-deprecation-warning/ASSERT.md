---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains the expected codex events.
- stderr does NOT contain any deprecation warning.

```go
import (
	"testing"
	"strings"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"clean codex"`)
    if strings.Contains(resp.Stderr, "deprecat") || strings.Contains(resp.Stderr, "stdout_events") {
        t.Fatalf("unexpected deprecation warning in stderr:\n%s", resp.Stderr)
    }
}
```
