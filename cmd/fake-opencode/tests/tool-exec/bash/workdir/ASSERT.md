## Expected
- The `pwd` output contains the expected workdir path.

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
    if !strings.Contains(output, "subdir") {
        t.Fatalf("expected pwd output to contain 'subdir', got: %q", output)
    }
}
```
