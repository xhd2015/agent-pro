## Expected
- `resp.Runner` is `"crush"` (direct parent detection).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "crush" {
        t.Fatalf("expected runner 'crush', got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
