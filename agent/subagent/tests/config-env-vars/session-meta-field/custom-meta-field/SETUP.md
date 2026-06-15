## Preconditions
- `Config.SessionMetaField` is set to `"my_custom_meta_field"`.
- A session exists with meta.json containing `"my_custom_meta_field": "target_session_value"`.
- The session env var resolves to `"target_session_value"`.

## Steps
1. Set `req.SessionMetaField = "my_custom_meta_field"`.
2. Set `req.SessionEnvVar = "MY_CUSTOM_ENV"` and `MY_CUSTOM_ENV=target_session_value` in env.
3. Pre-create a session with meta.json `{"my_custom_meta_field": "target_session_value", ...}`.
4. Call `Run()` with `Status: true`.
5. Verify the session is found (status displayed, no "session not found" error).

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "15", "sess_meta1")

    req.HomeDir = baseDir
    req.SessionBase = baseDir
    req.SessionID = ""
    req.SessionEnvVar = "MY_CUSTOM_ENV"
    req.SessionMetaField = "my_custom_meta_field"
    req.Env = append(req.Env,
        "MY_CUSTOM_ENV=target_session_value",
    )
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"explicit_meta_1","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z","my_custom_meta_field":"target_session_value"}`,
    }
    return nil
}
```
