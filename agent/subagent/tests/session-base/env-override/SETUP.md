## Preconditions
- `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` env var is set.
- `Options.SessionBase` is also set but should be overridden by the env var.

## Steps
1. Set `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` to a debug temp dir.
2. Set `req.SessionBase` to a different value.
3. Pre-create a session directory under the debug dir (not the SessionBase dir).
4. Verify the session is found from the debug dir (env override wins).

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    debugDir := filepath.Join(t.TempDir(), "debug_home")
    customDir := filepath.Join(t.TempDir(), "custom_should_be_ignored")
    sessDir := filepath.Join(debugDir, "sess_debug123")

    req.SessionBase = customDir
    req.Env = append(req.Env, "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME="+debugDir)
    req.PreCreateDirs = append(req.PreCreateDirs, sessDir)
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"debug_test_123","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    return nil
}
```
