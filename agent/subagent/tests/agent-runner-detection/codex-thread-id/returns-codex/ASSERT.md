## Expected
- `resp.Runner` is `"codex"`.
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "codex" {
        t.Fatalf("expected runner 'codex', got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
