## Preconditions
- `Options.SessionBase` is set to a custom directory.
- Session directories exist under that custom base.

## Steps
1. Set `req.SessionBase` to a temp dir.
2. Pre-create a session directory under `<SessionBase>/testrole/sessions/` with a `meta.json`.
3. Verify the session is listed from the custom base.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sessDir := filepath.Join(baseDir, "testrole", "sessions", "sess_test123")

    req.SessionBase = filepath.Join(baseDir, "testrole", "sessions")
    req.PreCreateDirs = append(req.PreCreateDirs, sessDir)
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"explicit_test_123","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    return nil
}
```
