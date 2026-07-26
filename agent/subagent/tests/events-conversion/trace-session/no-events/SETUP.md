## Preconditions
- No events.jsonl file exists in the session directory.

## Steps
1. Set up a session dir with no events.jsonl.
2. Call `traceSession`.
3. Verify output shows "(no events yet)".

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-no-events","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.SessionID = "test-no-events"
    req.SessionBase = dir
    return nil
}
```
