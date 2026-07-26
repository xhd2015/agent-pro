## Expected
- A non-nil error is returned indicating the runner is unknown.

```go
import (
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err == nil {
        t.Fatalf("expected error for unknown runner, got nil")
    }
}
```
