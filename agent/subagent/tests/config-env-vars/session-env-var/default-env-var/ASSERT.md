## Expected
- Stderr references `my_sid_default` (from the default env var).
- The default env var `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID` was used for resolution.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "my_sid_default") {
        t.Fatalf("expected 'my_sid_default' in stderr (default env var used), got:\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr, got:\n%s", resp.Stderr)
    }
}
```
