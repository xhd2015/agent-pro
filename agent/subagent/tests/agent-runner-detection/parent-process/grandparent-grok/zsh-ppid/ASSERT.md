## Expected
- `resp.Runner` is `"grok"` (detected via grandparent walk from zsh shell).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "grok" {
        t.Fatalf("expected runner 'grok' (grandparent walk via zsh), got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
