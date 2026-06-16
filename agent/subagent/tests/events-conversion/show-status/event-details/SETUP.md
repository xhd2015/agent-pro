## Preconditions
- events.jsonl exists with multiple AgentEvent lines (more than 3).

## Steps
1. Set up a session dir with events.jsonl containing 4 AgentEvent lines.
2. Call `showStatus`.
3. Verify the last events are shown with summaries.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    eventsContent := `{"type":"message","text":"event1","timestamp":1718444401000}
{"type":"think","text":"event2","timestamp":1718444402000}
{"type":"tool_call","tool":"bash","tool_input":{"command":"ls"},"output":"event3","timestamp":1718444403000}
{"type":"message","text":"event4 final","timestamp":1718444404000}
`

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-event-details","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-event-details"
    req.SessionBase = dir
    return nil
}
```
