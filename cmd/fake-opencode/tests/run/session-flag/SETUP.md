## Preconditions
- The command passes `--session sess_arg`.

## Steps
1. Run fake opencode with a mock config that does not define a session.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    sessDir := filepath.Join(req.TempDir, "opencode-config", "sessions", "sess_arg")
    if err := os.MkdirAll(sessDir, 0755); err != nil {
        return err
    }
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","llm_events":[{"type":"message","text":"session flag"}]}`)
    req.Args = []string{"run", "--format", "json", "--session", "sess_arg", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```
