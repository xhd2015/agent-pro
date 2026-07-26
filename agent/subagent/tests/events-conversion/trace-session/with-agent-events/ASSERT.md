## Expected
- Output contains formatted event information for the AgentEvent lines.
- The important text from events appears in output.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Should contain something about the events
    if !strings.Contains(resp.Stdout, "3 lines") {
        t.Fatalf("expected event count 3 in output, got:\n%s", resp.Stdout)
    }
    // Should contain formatted content from at least one event
    if !strings.Contains(resp.Stdout, "Here is the result") {
        t.Fatalf("expected message text in formatted output, got:\n%s", resp.Stdout)
    }
}
```
