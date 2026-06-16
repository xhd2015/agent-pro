## Expected
- `resp.Duration` equals `5m`.
- `resp.Err` is nil.
- `resp.Stderr` contains warning (since 5m < 10m).

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
