---
label: e2e
---

## Expected
- The event output contains the real file content: `"hello file content for read test"`.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    events := parseJSONLines(t, resp.Stdout)
    if len(events) == 0 {
        t.Fatal("no events in stdout")
    }
    event := events[0]
    part, _ := event["part"].(map[string]any)
    state, _ := part["state"].(map[string]any)
    output, _ := state["output"].(string)
    if !strings.Contains(output, "hello file content for read test") {
        t.Fatalf("expected file content in output, got: %q", output)
    }
    if status, _ := state["status"].(string); status != "completed" {
        t.Fatalf("expected status 'completed', got: %q", status)
    }
}
```
