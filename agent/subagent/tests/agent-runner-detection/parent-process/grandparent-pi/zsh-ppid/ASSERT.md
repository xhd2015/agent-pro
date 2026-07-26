## Expected
- `resp.Runner` is `"pi"` (detected via grandparent walk from zsh shell).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "pi" {
        t.Fatalf("expected runner 'pi' (grandparent walk via zsh), got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
