## Preconditions
- A session exists with `events.jsonl` containing formatted events.
- The session is not live (no pid file).

## Steps
1. Pre-create a session with `meta.json` and `events.jsonl`.
2. Set `req.SessionID` to match the session.
3. Do NOT create a pid file (finished session).
4. Call `traceSession`.
5. Verify formatted event lines appear in output.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "15", "sess_trace2")

    req.HomeDir = baseDir
    req.SessionBase = baseDir
    req.SessionID = "trace_events_1"
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"trace_events_1","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: `{"type":"tool_use","tool":"bash","timestamp":1718444400000,"delta":{"command":"echo hello"}}` + "\n" +
            `{"type":"tool_result","content":"hello","timestamp":1718444401000}` + "\n",
    }
    req.PreCreatePID = false
    return nil
}
```
