## Preconditions
- A session exists but has no `events.jsonl` file.

## Steps
1. Pre-create a session directory with `meta.json` but no `events.jsonl`.
2. Set `req.SessionID` to match the session.
3. Write a pid file so the session appears live (to avoid premature termination).
4. Call `traceSession`.
5. Verify "(no events yet)" message appears.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "15", "sess_trace1")

    req.HomeDir = baseDir
    req.SessionBase = baseDir
    req.SessionID = "trace_noevents_1"
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"trace_noevents_1","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    req.PreCreatePID = true
    return nil
}
```
