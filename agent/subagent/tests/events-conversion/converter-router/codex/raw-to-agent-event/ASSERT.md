## Expected
- At least one AgentEvent with `"type":"tool_call"` (from command_execution).
- At least one AgentEvent with `"type":"message"` and text "Codex says hello".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(resp.Stdout, `"type":"tool_call"`) {
        t.Fatalf("expected tool_call event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"type":"message"`) {
        t.Fatalf("expected message event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "Codex says hello") {
        t.Fatalf("expected message text, got:\n%s", resp.Stdout)
    }
}
```
