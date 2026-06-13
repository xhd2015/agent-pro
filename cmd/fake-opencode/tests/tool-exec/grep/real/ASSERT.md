## Expected
- The event output contains `"UNIQUE_MARKER_FOR_GREP"` — the real grep result.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    events := parseJSONLines(t, resp.Stdout)
    if len(events) == 0 {
        t.Fatal("no events in stdout")
    }
    event := events[0]
    part, _ := event["part"].(map[string]any)
    state, _ := part["state"].(map[string]any)
    output, _ := state["output"].(string)
    if !strings.Contains(output, "UNIQUE_MARKER_FOR_GREP") {
        t.Fatalf("expected grep output containing 'UNIQUE_MARKER_FOR_GREP', got: %q", output)
    }
}
```
