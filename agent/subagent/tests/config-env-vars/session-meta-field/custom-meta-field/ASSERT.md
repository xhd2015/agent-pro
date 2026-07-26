## Expected
- Stderr does NOT contain "session not found".
- Stdout contains the session status information (session found via custom meta field).
- The session matched by the custom `SessionMetaField`.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected session to be found via custom meta field, but got 'session not found':\n%s", resp.Stderr)
    }

    if !strings.Contains(resp.Stdout, "Session") {
        t.Fatalf("expected session status in stdout, got:\n%s", resp.Stdout)
    }
}
```
