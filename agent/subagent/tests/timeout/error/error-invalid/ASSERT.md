## Expected
- `resp.Err` is not nil.
- Error is a parse error (not a minimum-duration error).

```go
import (
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if resp.Err == nil {
        t.Fatal("expected error for invalid input 'abc', got nil")
    }
}
```
