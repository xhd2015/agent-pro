## Preconditions
- `Config.DebugSessionEnv` is empty (default behavior).
- The default env var `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` is set.

## Steps
1. Leave `req.DebugSessionEnv` empty.
2. Create a directory with session data.
3. Set `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME=<dir>` in env.
4. Call `Run()` with `ListSessions: true`.
5. Verify sessions are listed from the default debug env dir.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    debugDir := filepath.Join(t.TempDir(), "debug_home_default")
    sessDir := filepath.Join(debugDir, "sess_debug2")

    req.DebugSessionEnv = ""
    req.Env = append(req.Env,
        "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME="+debugDir,
    )
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"debug_default_session","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    return nil
}
```
