---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains three JSON lines with reasoning, tool_use, and text events in order.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    lines := strings.Split(strings.TrimSpace(resp.Stdout), "\n")
    if len(lines) != 3 {
        t.Fatalf("expected 3 JSON lines, got %d\n%s", len(lines), resp.Stdout)
    }
    assertContains(t, lines[0], `"type":"reasoning"`)
    assertContains(t, lines[0], `"initial analysis"`)
    assertContains(t, lines[1], `"type":"tool_use"`)
    assertContains(t, lines[1], `"mid command"`)
    assertContains(t, lines[2], `"type":"text"`)
    assertContains(t, lines[2], `"final summary"`)
}
```
