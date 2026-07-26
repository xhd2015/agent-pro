---
label: e2e
---

## Expected
- The tool_use event contains `"output":"hello real bash\n"` (or similar) — the real stdout.
- The event status is `"completed"`.
- The exit code is 0.

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
    part, ok := event["part"].(map[string]any)
    if !ok {
        t.Fatalf("event has no part object: %v", event)
    }
    state, ok := part["state"].(map[string]any)
    if !ok {
        t.Fatalf("part has no state object: %v", part)
    }
    output, _ := state["output"].(string)
    if !strings.Contains(output, "hello real bash") {
        t.Fatalf("expected real bash output containing 'hello real bash', got: %q", output)
    }
    if status, _ := state["status"].(string); status != "completed" {
        t.Fatalf("expected status 'completed', got: %q", status)
    }
    exitCode, hasExit := state["exit_code"]
    if hasExit {
        if code, ok := exitCode.(float64); ok && int(code) != 0 {
            t.Fatalf("expected exit_code 0, got: %v", exitCode)
        }
    }
}
```
