## Expected
- Output header contains the session ID and event count.
- Formatted event lines appear in output.
- "Session finished" or "Done" message appears at the end.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "trace_events_1") {
        t.Fatalf("expected 'trace_events_1' in stdout, got:\n%s", resp.Stdout)
    }

    hasContent := strings.Contains(resp.Stdout, "echo") || strings.Contains(resp.Stdout, "hello") || strings.Contains(resp.Stdout, "tool")
    if !hasContent {
        t.Fatalf("expected event content in stdout, got:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "Events") && !strings.Contains(resp.Stdout, "lines") {
        t.Fatalf("expected event metadata in stdout, got:\n%s", resp.Stdout)
    }
}
```
