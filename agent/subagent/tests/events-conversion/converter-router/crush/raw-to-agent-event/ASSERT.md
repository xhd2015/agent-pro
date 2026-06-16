## Expected
- At least one AgentEvent with `"type":"message"` and text "Crush response".
- At least one AgentEvent with `"type":"tool_call"` and tool "read".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(resp.Stdout, `"type":"message"`) {
        t.Fatalf("expected message event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "Crush response") {
        t.Fatalf("expected message text, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"type":"tool_call"`) {
        t.Fatalf("expected tool_call event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"tool":"read"`) {
        t.Fatalf("expected tool read, got:\n%s", resp.Stdout)
    }
}
```
