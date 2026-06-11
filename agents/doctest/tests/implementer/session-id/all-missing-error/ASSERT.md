## Expected
- When NOT running under opencode: exit code non-zero, stderr lists all 3 options.
- When running under opencode: auto-discovery succeeds, test skips.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Skip("auto-discovery succeeded (running under opencode)")
    }

    if !strings.Contains(resp.Stderr, "must be run from codex or opencode") {
        t.Fatalf("stderr missing 'must be run from codex or opencode':\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "DOCTEST_AGENT_IMPLEMENTER_SESSION_ID") {
        t.Fatalf("stderr missing 'DOCTEST_AGENT_IMPLEMENTER_SESSION_ID':\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "CODEX_THREAD_ID") {
        t.Fatalf("stderr missing 'CODEX_THREAD_ID':\n%s", resp.Stderr)
    }
}
```
