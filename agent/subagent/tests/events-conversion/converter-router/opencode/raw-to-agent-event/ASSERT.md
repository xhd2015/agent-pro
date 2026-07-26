## Expected
- At least one AgentEvent with `"type":"think"` and text "analyzing the request".
- At least one AgentEvent with `"type":"tool_call"`, tool "bash", and tool_input containing `"command":"ls"`.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(resp.Stdout, `"type":"think"`) {
        t.Fatalf("expected think event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "analyzing the request") {
        t.Fatalf("expected think text, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"type":"tool_call"`) {
        t.Fatalf("expected tool_call event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"tool":"bash"`) {
        t.Fatalf("expected tool bash, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `"command":"ls"`) {
        t.Fatalf("expected command:ls in tool_input, got:\n%s", resp.Stdout)
    }
}
```
