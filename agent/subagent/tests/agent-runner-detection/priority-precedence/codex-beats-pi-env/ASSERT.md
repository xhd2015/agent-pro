## Expected
- `resp.Runner` is `"codex"` (P2 CODEX_THREAD_ID beats P3 PI_CODING_AGENT).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "codex" {
        t.Fatalf("expected runner 'codex' (P2 beats P3), got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
