## Expected
- Formatted output contains the message text.
- The output is non-empty.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Stdout == "" {
        t.Fatalf("expected non-empty formatted output, got empty string")
    }
    if !strings.Contains(resp.Stdout, "Timestamped message") {
        t.Fatalf("expected message in output, got:\n%s", resp.Stdout)
    }
}
```
