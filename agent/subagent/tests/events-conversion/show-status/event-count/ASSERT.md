## Expected
- Output shows the correct event count (5 lines).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(resp.Stdout, "5 lines") {
        t.Fatalf("expected event count 5 in output, got:\n%s", resp.Stdout)
    }
}
```
