## Expected
- `Run` returns no "cannot detect session id" error (`resp.Err == nil`).
- Stderr contains `session not found` (resolution succeeded with a generated ID; session directory does not exist yet).
- Stderr does not contain `cannot detect session id`.

## Side Effects
- None (no session directory created).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.Err != nil {
        t.Fatalf("expected no error from Run, got: %v", resp.Err)
    }

    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr (generated ID used for lookup), got:\n%s", resp.Stderr)
    }

    if strings.Contains(resp.Stderr, "cannot detect session id") {
        t.Fatalf("expected no 'cannot detect session id' in stderr, got:\n%s", resp.Stderr)
    }
}
```