## Expected
- `resp.Duration` equals `5m` (300_000_000_000 nanoseconds).
- `resp.Err` is nil.
- `resp.Stderr` contains a warning about suggesting a longer timeout.

```go
import (
    "strings"
    "testing"
    "time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    expected := 5 * time.Minute
    if resp.Duration != expected {
        t.Fatalf("expected duration %v (5m), got %v", expected, resp.Duration)
    }

    if resp.Err != nil {
        t.Fatalf("expected no error, got: %v", resp.Err)
    }

    if !strings.Contains(resp.Stderr, "longer") && !strings.Contains(resp.Stderr, "suggest") {
        t.Fatalf("expected warning about short timeout in stderr, got: %q", resp.Stderr)
    }
}
```
