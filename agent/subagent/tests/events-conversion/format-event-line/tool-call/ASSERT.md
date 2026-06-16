## Expected
- Formatted output contains tool-call indication with the tool name (e.g., "bash" or "RUN").
- The output is non-empty.

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
        t.Fatalf("expected non-empty formatted output, got empty string")
    }
    lower := strings.ToLower(resp.Stdout)
    if !strings.Contains(lower, "bash") && !strings.Contains(lower, "run") {
        t.Fatalf("expected 'bash' or 'RUN' in formatted output, got:\n%s", resp.Stdout)
    }
}
```
