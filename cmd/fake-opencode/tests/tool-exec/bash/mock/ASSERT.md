## Expected
- The output is `"fake mock output"` — the mock value.
- The output does **not** contain `"REAL OUTPUT"` (proving execution was skipped).

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
    if !strings.Contains(output, "fake mock output") {
        t.Fatalf("expected mock output 'fake mock output', got: %q", output)
    }
    if strings.Contains(output, "REAL OUTPUT") {
        t.Fatalf("mock should prevent real execution, but got real output: %q", output)
    }
}
```
