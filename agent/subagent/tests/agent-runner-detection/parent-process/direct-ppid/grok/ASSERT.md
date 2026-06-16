## Expected
- `resp.Runner` is `"grok"` (direct parent detection).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "grok" {
        t.Fatalf("expected runner 'grok', got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
