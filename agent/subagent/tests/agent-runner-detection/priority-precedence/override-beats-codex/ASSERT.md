## Expected
- `resp.Runner` is `"pi"` (env override at P1 beats CODEX_THREAD_ID at P2).
- `resp.Detected` is `true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Runner != "pi" {
        t.Fatalf("expected runner 'pi' (P1 override wins), got %q", resp.Runner)
    }
    if !resp.Detected {
        t.Fatalf("expected detected=true, got false")
    }
}
```
