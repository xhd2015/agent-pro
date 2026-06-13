## Expected
- The command succeeds.
- stdout contains all seven opencode event types and the expected text/content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    lines := parseJSONLines(t, resp.Stdout)
    types := map[string]bool{}
    for _, line := range lines {
        if t, ok := line["type"].(string); ok {
            types[t] = true
        }
    }
    if !types["reasoning"] {
        t.Fatalf("missing reasoning event in:\n%s", resp.Stdout)
    }
    if !types["tool_use"] {
        t.Fatalf("missing tool_use event in:\n%s", resp.Stdout)
    }
    if !types["text"] {
        t.Fatalf("missing text event in:\n%s", resp.Stdout)
    }
    if !types["error"] {
        t.Fatalf("missing error event in:\n%s", resp.Stdout)
    }
    if !types["done"] {
        t.Fatalf("missing done event in:\n%s", resp.Stdout)
    }
    if !types["step_start"] {
        t.Fatalf("missing step_start event in:\n%s", resp.Stdout)
    }
    if !types["step_finish"] {
        t.Fatalf("missing step_finish event in:\n%s", resp.Stdout)
    }
}
```
