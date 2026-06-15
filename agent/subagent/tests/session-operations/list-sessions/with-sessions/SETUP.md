## Preconditions
- Multiple session directories exist under the configured base with different creation times.

## Steps
1. Pre-create two session directories with different timestamps.
2. Call `ListSessions`.
3. Verify sessions are listed sorted by creation time descending.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    baseDir := filepath.Join(t.TempDir(), "custom_base")
    sess1 := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "15", "sess_sessA")
    sess2 := filepath.Join(baseDir, "testrole", "sessions", "2026", "06", "14", "sess_sessB")

    req.HomeDir = baseDir
    req.SessionBase = baseDir
    req.PreCreateDirs = []string{sess1, sess2}
    req.PreCreateMeta = map[string]string{
        sess1: `{"explicit_session_id":"session_alpha","agent_runner":"opencode","created_at":"2026-06-15T10:00:00Z"}`,
        sess2: `{"explicit_session_id":"session_beta","agent_runner":"opencode","created_at":"2026-06-14T10:00:00Z"}`,
    }
    return nil
}
```
