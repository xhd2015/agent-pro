## Expected
- The event output contains `"fake read content"` (the mock value).
- The event output does **not** contain `"this should not appear"` (the real file content).

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
    if !strings.Contains(output, "fake read content") {
        t.Fatalf("expected mock content 'fake read content', got: %q", output)
    }
    if strings.Contains(output, "this should not appear") {
        t.Fatalf("mock should prevent real file read, got real content: %q", output)
    }
}
```
