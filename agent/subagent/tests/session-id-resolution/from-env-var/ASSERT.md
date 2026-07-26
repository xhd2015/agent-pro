## Expected
- An error is printed to stderr indicating session not found for `env_session_123`.
- The env var session ID was used for lookup (not a generated ID).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "env_session_123") {
        t.Fatalf("expected session ID 'env_session_123' in stderr, got:\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr, got:\n%s", resp.Stderr)
    }
}
```
