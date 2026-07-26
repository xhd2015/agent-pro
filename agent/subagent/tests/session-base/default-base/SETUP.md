## Preconditions
- `Options.SessionBase` is empty.
- The default base is `~/.agent-pro/subagent/<role>/sessions/`.

## Steps
1. Set `req.HomeDir` to a temp dir so `~` resolves predictably.
2. Leave `req.SessionBase` empty (default behavior).
3. Pre-create a session directory under `$HOME/.agent-pro/subagent/testrole/sessions/` with a `meta.json`.
4. Verify the session is listed from the default base.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    home := t.TempDir()
    sessDir := filepath.Join(home, ".agent-pro", "subagent", "testrole", "sessions", "sess_default123")

    req.HomeDir = home
    req.SessionBase = ""
    req.PreCreateDirs = append(req.PreCreateDirs, sessDir)
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"default_test_123","agent_runner":"opencode","created_at":"2026-06-15T12:00:00Z"}`,
    }
    return nil
}
```
