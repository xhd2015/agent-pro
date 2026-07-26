## Expected
- `resp.Err` is not nil (30s < 1m).

```go
import (
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if resp.Err == nil {
        t.Fatal("expected error for suffix seconds < 1m, got nil")
    }
}
```
