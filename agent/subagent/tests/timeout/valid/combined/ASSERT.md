## Expected
- `resp.Duration` equals `1h30m` (90 minutes).
- `resp.Err` is nil.
- `resp.Stderr` is empty.

```go
import (
    "testing"
    "time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    expected := time.Hour + 30*time.Minute
    if resp.Duration != expected {
        t.Fatalf("expected duration %v (1h30m), got %v", expected, resp.Duration)
    }

    if resp.Err != nil {
        t.Fatalf("expected no error, got: %v", resp.Err)
    }

    if resp.Stderr != "" {
        t.Fatalf("expected no stderr output, got: %q", resp.Stderr)
    }
}
```
