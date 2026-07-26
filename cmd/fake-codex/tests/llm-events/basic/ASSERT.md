---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains codex reasoning events (item.started and item.completed).
- The reasoning text appears in the completed event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"thinking in codex"`)
}
```
