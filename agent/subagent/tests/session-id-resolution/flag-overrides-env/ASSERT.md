## Expected
- Stderr references `flag_session_456` (the flag value), not `env_session_123`.
- The flag takes priority over the env var.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "flag_session_456") {
        t.Fatalf("expected flag session ID 'flag_session_456' in stderr, got:\n%s", resp.Stderr)
    }
    if strings.Contains(resp.Stderr, "env_session_123") {
        t.Fatalf("expected env var session ID 'env_session_123' NOT in stderr, but found:\n%s", resp.Stderr)
    }
}
```
