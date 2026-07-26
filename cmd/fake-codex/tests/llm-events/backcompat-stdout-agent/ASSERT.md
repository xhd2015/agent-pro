---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains codex reasoning events with the expected text.
- stderr contains deprecation warning.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"agent format backcompat"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
