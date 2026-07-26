## Preconditions
- events.jsonl exists with multiple AgentEvent lines.

## Steps
1. Set up a session dir with events.jsonl containing 5 AgentEvent lines.
2. Call `showStatus`.
3. Verify event count is shown correctly.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    eventsContent := `{"type":"message","text":"one","timestamp":1718444401000}
{"type":"think","text":"two","timestamp":1718444402000}
{"type":"tool_call","tool":"bash","output":"three","timestamp":1718444403000}
{"type":"message","text":"four","timestamp":1718444404000}
{"type":"message","text":"five","timestamp":1718444405000}
`

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-event-count","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-event-count"
    req.SessionBase = dir
    return nil
}
```
