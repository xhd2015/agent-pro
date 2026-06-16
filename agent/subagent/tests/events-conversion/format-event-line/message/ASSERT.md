## Expected
- Formatted output contains the message text "Hello, this is a test message".
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
    if !strings.Contains(resp.Stdout, "Hello, this is a test message") {
        t.Fatalf("expected message text in output, got:\n%s", resp.Stdout)
    }
}
```
