## Expected
- Output contains a relative time indicator like "ago" or "s".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Should contain relative time indicator
    if !strings.Contains(resp.Stdout, "ago") && !strings.Contains(resp.Stdout, "0s") && !strings.Contains(resp.Stdout, "1s") {
        t.Fatalf("expected relative time in output, got:\n%s", resp.Stdout)
    }
}
```
