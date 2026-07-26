## Preconditions
- A session exists with a pid file containing the current process PID.
- The process is still alive.

## Steps
1. Pre-create a session with meta.json.
2. Pre-create a pid file with current PID (re.PreCreatePID = true).
3. Set `req.SessionID` to match the session.
4. Call `showStatus`.
5. Verify status shows "running".

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "sess_running1")

    req.HomeDir = baseDir
    req.SessionBase = filepath.Join(baseDir, "testrole", "sessions")
    req.SessionID = "running_session_1"
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"running_session_1","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    req.PreCreatePID = true
    return nil
}
```
