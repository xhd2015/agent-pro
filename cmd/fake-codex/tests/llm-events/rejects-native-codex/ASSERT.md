---
label: e2e
---

## Expected
- The command fails with a non-zero exit code (unrecognized event type in llm_events).
- Alternatively, if unrecognized types are silently skipped, stdout must NOT contain native codex output.

```go
import (
	"testing"
	"strings"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode != 0 && strings.Contains(resp.Stderr, "llm_events") {
        return
    }
    // If it succeeds, verify no native codex output leaked
    assertNotContains(t, resp.Stdout, `"item.started"`)
    assertNotContains(t, resp.Stdout, `"native format"`)
}
```
