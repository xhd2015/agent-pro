---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains the native codex event (item.started directly, no conversion needed).
- stderr contains deprecation warning.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"native format backcompat"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
