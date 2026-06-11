## Expected
- Exit code non-zero.
- Stderr contains "CODEX_THREAD_ID must be set".

## Exit Code
- Exit code non-zero.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit for missing CODEX_THREAD_ID")
    }
    if !strings.Contains(resp.Stderr, "CODEX_THREAD_ID must be set") {
        t.Fatalf("expected 'CODEX_THREAD_ID must be set' in stderr, got:\n%s", resp.Stderr)
    }
}
```
