## Expected
- Formatted output is non-empty for step_start event.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Stdout == "" {
        t.Fatalf("expected non-empty formatted output for step_start, got empty string")
    }
    _ = strings.ToLower(resp.Stdout) // at least it rendered something
}
```
