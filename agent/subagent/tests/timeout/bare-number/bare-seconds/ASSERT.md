## Expected
- `resp.Err` is not nil (duration < 1m).
- Error message mentions minimum duration requirement.

```go
import (
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if resp.Err == nil {
        t.Fatal("expected error for bare seconds < 1m, got nil")
    }
}
```
