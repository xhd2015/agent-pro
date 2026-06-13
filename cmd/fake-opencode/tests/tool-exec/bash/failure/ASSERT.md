## Expected
- The tool_use event has `exit_code` 3 (or non-zero).
- The overall fake-opencode process still exits 0 (tool failures don't cause runner to fail).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
    exitCode, ok := state["exit_code"]
    if !ok {
        t.Fatal("expected exit_code in state, not found")
    }
    code, ok := exitCode.(float64)
    if !ok {
        t.Fatalf("exit_code is not a number: %T %v", exitCode, exitCode)
    }
    if int(code) == 0 {
        t.Fatalf("expected non-zero exit_code, got 0")
    }
}
```
