## Expected
- `resp.Runner` is `""` (no agent detected).
- `resp.Detected` is `false`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "" {
        t.Fatalf("expected runner '', got %q", resp.Runner)
    }
    if resp.Detected {
        t.Fatalf("expected detected=false, got true")
    }
}
```
