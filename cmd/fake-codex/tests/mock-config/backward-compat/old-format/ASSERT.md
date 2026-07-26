---
label: e2e
---

## Expected
- The command succeeds.
- The legacy codex event is emitted exactly as configured.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"legacy format ok"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"type":"message"`)
}
```
