## Preconditions
- events.jsonl exists with an AgentEvent that has a recent timestamp.

## Steps
1. Set up a session dir with an event that has a timestamp ~1 second ago.
2. Call `showStatus`.
3. Verify "1s ago" or similar relative time appears.

```go
import (
    "fmt"
    "path/filepath"
    "testing"
    "time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    ts := time.Now().UnixMilli()
    eventsContent := fmt.Sprintf(`{"type":"message","text":"recent","timestamp":%d}
`, ts)
    _ = ts

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-last-event-time","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-last-event-time"
    req.SessionBase = dir
    return nil
}
```
