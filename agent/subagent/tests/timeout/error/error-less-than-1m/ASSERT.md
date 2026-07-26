## Expected
- `resp.Err` is not nil.
- Error message contains "at least 1m" or "minimum" or "1m".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if resp.Err == nil {
        t.Fatal("expected error for 30s < 1m, got nil")
    }

    msg := resp.Err.Error()
    if !strings.Contains(msg, "1m") && !strings.Contains(msg, "minimum") && !strings.Contains(msg, "least") {
        t.Fatalf("expected error to mention minimum duration, got: %v", resp.Err)
    }
}
```
