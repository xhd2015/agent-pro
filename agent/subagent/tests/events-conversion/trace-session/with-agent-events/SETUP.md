## Preconditions
- events.jsonl exists with AgentEvent JSON lines.

## Steps
1. Set up a session dir with events.jsonl containing AgentEvent lines.
2. Call `traceSession`.
3. Verify formatted output appears for each event.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    eventsContent := `{"type":"think","text":"Let me reason about this"}
{"type":"tool_call","tool":"bash","tool_input":{"command":"ls"},"output":"file1.txt"}
{"type":"message","text":"Here is the result"}
`

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-with-events","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-with-events"
    req.SessionBase = dir
    return nil
}
```
