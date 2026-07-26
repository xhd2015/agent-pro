---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains all three event types in order.
- The JSON lines appear as: reasoning events, then command_execution events, then message event.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    lines := strings.Split(strings.TrimSpace(resp.Stdout), "\n")
    if len(lines) == 0 {
        t.Fatal("no JSON lines emitted")
    }
    combined := strings.Join(lines, " ")
    assertContains(t, combined, `"type":"reasoning"`)
    assertContains(t, combined, `"step 1 analyze"`)
    assertContains(t, combined, `"type":"command_execution"`)
    assertContains(t, combined, `"step 2 execute"`)
    assertContains(t, combined, `"type":"message"`)
    assertContains(t, combined, `"step 3 done"`)
}
```
