## Preconditions
- `Config.DebugSessionEnv` is set to `"MY_DEBUG_HOME"`.
- The env var `MY_DEBUG_HOME` points to a directory containing session data.

## Steps
1. Set `req.DebugSessionEnv = "MY_DEBUG_HOME"`.
2. Create a custom debug home directory with session data.
3. Set `MY_DEBUG_HOME=<debug_dir>` in env.
4. Also set `req.SessionBase` to a different value (should be ignored because debug env overrides).
5. Call `Run()` with `ListSessions: true`.
6. Verify sessions are listed from the debug home dir.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    debugDir := filepath.Join(t.TempDir(), "debug_home_custom")
    ignoredDir := filepath.Join(t.TempDir(), "should_be_ignored")
    sessDir := filepath.Join(debugDir, "sess_debug1")

    req.DebugSessionEnv = "MY_DEBUG_HOME"
    req.SessionBase = ignoredDir
    req.Env = append(req.Env,
        "MY_DEBUG_HOME="+debugDir,
    )
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"debug_custom_session","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    return nil
}
```
