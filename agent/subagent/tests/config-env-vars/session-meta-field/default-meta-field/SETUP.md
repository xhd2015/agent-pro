## Preconditions
- `Config.SessionMetaField` is empty (default behavior).
- The default meta field `"subagent_role_testrole_session_id"` is used.
- A session exists with meta.json containing the default field name.

## Steps
1. Leave `req.SessionMetaField` empty.
2. Set `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=target_default_meta` in env.
3. Pre-create a session with meta.json `{"subagent_role_testrole_session_id": "target_default_meta", ...}`.
4. Call `Run()` with `Status: true`.
5. Verify the session is found via the default meta field.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "sess_meta2")

    req.HomeDir = baseDir
    req.SessionBase = filepath.Join(baseDir, "testrole", "sessions")
    req.SessionID = ""
    req.SessionEnvVar = ""
    req.SessionMetaField = ""
    req.Env = append(req.Env,
        "AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=target_default_meta",
    )
    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"explicit_meta_2","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z","subagent_role_testrole_session_id":"target_default_meta"}`,
    }
    return nil
}
```
