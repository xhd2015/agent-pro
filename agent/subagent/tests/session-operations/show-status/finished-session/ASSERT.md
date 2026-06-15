## Expected
- Status output shows the session ID, runner, created time.
- Status is "finished" (or not "running").
- Event count reflects the events.jsonl content.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "finished_session_1") {
        t.Fatalf("expected 'finished_session_1' in stdout, got:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "opencode") {
        t.Fatalf("expected 'opencode' runner in stdout, got:\n%s", resp.Stdout)
    }

    if strings.Contains(resp.Stdout, "running") {
        t.Fatalf("expected finished session, got 'running' in:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "Events") {
        t.Fatalf("expected 'Events' section in stdout, got:\n%s", resp.Stdout)
    }
}
```
