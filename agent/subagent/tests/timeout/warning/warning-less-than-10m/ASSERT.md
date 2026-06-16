## Expected
- `resp.Duration` equals `3m`.
- `resp.Err` is nil.
- `resp.Stderr` contains a warning suggesting a longer timeout (e.g., `--timeout=1h`).

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

    expected := 3 * time.Minute
    if resp.Duration != expected {
        t.Fatalf("expected duration %v (3m), got %v", expected, resp.Duration)
    }

    if resp.Err != nil {
        t.Fatalf("expected no error, got: %v", resp.Err)
    }

    if !strings.Contains(resp.Stderr, "longer") && !strings.Contains(resp.Stderr, "suggest") {
        t.Fatalf("expected warning about short timeout in stderr, got: %q", resp.Stderr)
    }
}
```
