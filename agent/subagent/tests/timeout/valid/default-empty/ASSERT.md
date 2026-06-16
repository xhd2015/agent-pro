## Expected
- `resp.Duration` equals `1h` (3_600_000_000_000 nanoseconds).
- `resp.Err` is nil.
- `resp.Stderr` is empty.

## Side Effects
- None.

```go
import (
    "testing"
    "time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    expected := time.Hour
    if resp.Duration != expected {
        t.Fatalf("expected duration %v (1h), got %v", expected, resp.Duration)
    }

    if resp.Err != nil {
        t.Fatalf("expected no error, got: %v", resp.Err)
    }

    if resp.Stderr != "" {
        t.Fatalf("expected no stderr output, got: %q", resp.Stderr)
    }
}
```
