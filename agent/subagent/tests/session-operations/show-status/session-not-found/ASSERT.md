## Expected
- Stderr contains "session not found".
- The session ID being sought is mentioned.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr, got:\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "nonexistent_session") {
        t.Fatalf("expected 'nonexistent_session' in stderr, got:\n%s", resp.Stderr)
    }
}
```
