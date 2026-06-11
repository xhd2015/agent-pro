## Expected
- The command succeeds quickly.

```go
import (
    "testing"
    "time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.Duration > 2*time.Second {
        t.Fatalf("duration = %s, want under 2s", resp.Duration)
    }
}
```

