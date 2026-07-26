## Expected
- `resp.Runner` is `""` (opencode at grandparent is ignored — pi-only walk).
- `resp.Detected` is `false`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "" {
        t.Fatalf("expected runner '' (opencode grandparent NOT detected), got %q", resp.Runner)
    }
    if resp.Detected {
        t.Fatalf("expected detected=false (opencode at grandparent ignored), got true")
    }
}
```
