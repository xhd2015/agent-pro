## Preconditions
- A finished session exists with meta.json and events.jsonl.
- No pid file (session is finished).

## Steps
1. Pre-create a session with meta.json and some events.
2. Set `req.SessionID` to match the session.
3. Do NOT create a pid file (session finished).
4. Call `showStatus`.
5. Verify formatted status output shows "finished".

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "15", "sess_finished1")

    req.HomeDir = baseDir
    req.SessionBase = baseDir
    req.SessionID = "finished_session_1"
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"finished_session_1","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z","main_agent_codex_thread_id":"codex_1","opencode_session_id":"oc_1"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: `{"type":"message","content":"hello","timestamp":1718444400000}` + "\n",
    }
    req.PreCreatePID = false
    return nil
}
```
