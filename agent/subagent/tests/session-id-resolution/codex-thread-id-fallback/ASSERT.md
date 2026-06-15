## Expected
- Stderr references `codex_thread_abc` as the session ID used for lookup.
- `CODEX_THREAD_ID` is used as the session ID fallback.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "codex_thread_abc") {
        t.Fatalf("expected CODEX_THREAD_ID 'codex_thread_abc' referenced in stderr, got:\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr, got:\n%s", resp.Stderr)
    }
}
```
