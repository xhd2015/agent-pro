## Expected
- `resp.Duration` equals `10m`.
- `resp.Err` is nil.
- `resp.Stderr` is empty (10m ≥ 10m threshold, no warning).

```go
import (
    "testing"
    "time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    expected := 10 * time.Minute
    if resp.Duration != expected {
        t.Fatalf("expected duration %v (10m), got %v", expected, resp.Duration)
    }

    if resp.Err != nil {
        t.Fatalf("expected no error, got: %v", resp.Err)
    }

    if resp.Stderr != "" {
        t.Fatalf("expected no stderr output at 10m boundary, got: %q", resp.Stderr)
    }
}
```
