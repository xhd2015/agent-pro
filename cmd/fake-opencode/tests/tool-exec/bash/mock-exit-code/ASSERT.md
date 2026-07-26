---
label: e2e
---

## Expected
- The mock exit_code 42 is present in the event.
- The mock stderr `"custom error message"` is present in the event.

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
    exitCode, ok := state["exit_code"]
    if !ok {
        t.Fatal("expected exit_code in state, not found")
    }
    code, ok := exitCode.(float64)
    if !ok {
        t.Fatalf("exit_code is not a number: %T %v", exitCode, exitCode)
    }
    if int(code) != 42 {
        t.Fatalf("expected mock exit_code 42, got: %v", int(code))
    }
    stderr, _ := state["stderr"].(string)
    if !strings.Contains(stderr, "custom error message") {
        t.Fatalf("expected mock stderr 'custom error message', got: %q", stderr)
    }
}
```
