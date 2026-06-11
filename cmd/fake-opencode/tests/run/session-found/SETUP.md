## Preconditions
- A session directory exists at `$OPENCODE_CONFIG_DIR/sessions/sess_known/`.
- The command passes `--session sess_known`.

## Steps
1. Pre-create the session directory.
2. Run fake opencode with the matching session flag.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    sessDir := filepath.Join(req.TempDir, "opencode-config", "sessions", "sess_known")
    if err := os.MkdirAll(sessDir, 0755); err != nil {
        return err
    }
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"session found no error"}}]}`)
    req.Args = []string{"run", "--format", "json", "--session", "sess_known", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```
