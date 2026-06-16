## Expected
- Formatted output contains the thinking text "Let me think about this problem carefully".
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
    if !strings.Contains(resp.Stdout, "Let me think about this problem carefully") {
        t.Fatalf("expected thinking text in output, got:\n%s", resp.Stdout)
    }
}
```
