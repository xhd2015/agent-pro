## Expected
- Status output shows "running" and the current PID.
- The session is live because the PID matches the current process.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "running_session_1") {
        t.Fatalf("expected 'running_session_1' in stdout, got:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "running") {
        t.Fatalf("expected 'running' status in stdout, got:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "PID") {
        t.Fatalf("expected 'PID' in stdout, got:\n%s", resp.Stdout)
    }
}
```
