## Expected
- Stderr does NOT contain "session not found".
- Stdout contains session status information.
- The default meta field `subagent_role_testrole_session_id` was used to match.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected session to be found via default meta field, but got 'session not found':\n%s", resp.Stderr)
    }

    if !strings.Contains(resp.Stdout, "Session") {
        t.Fatalf("expected session status in stdout, got:\n%s", resp.Stdout)
    }
}
```
