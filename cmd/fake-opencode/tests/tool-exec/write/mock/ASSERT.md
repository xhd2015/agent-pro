## Expected
- The event output contains `"mocked write done"`.
- The file was **not** created on disk (mock prevented real write).

```go
import (
    "os"
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
    if !strings.Contains(output, "mocked write done") {
        t.Fatalf("expected mock output 'mocked write done', got: %q", output)
    }

    // File should NOT have been created
    if _, statErr := os.Stat(req.TempDir + "/should-not-exist.txt"); statErr == nil {
        t.Fatal("mock should prevent real file write, but file was created")
    }
}
```
