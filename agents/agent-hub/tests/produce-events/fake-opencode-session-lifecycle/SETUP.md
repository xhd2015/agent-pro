## Preconditions
- A mock config defines session.created (before_stdout) and session.idle (before_exit) hooks.

## Steps
1. Create mock config with hooks.
2. Run `fake-opencode run --mock-config <path>`.
3. Verify 2 events emitted (started + finished) and session status is "completed".

```go
import (
    "testing"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    config := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_life","hooks":[{"at":"before_stdout","event":"session.created"},{"at":"before_exit","event":"session.idle"}]}`
    cfgPath := filepath.Join(req.TempDir, "mock-lifecycle.json")
    writeFile(t, cfgPath, config)
    req.Args = []string{"run", "--format", "json", "--mock-config", cfgPath, "hello"}
    req.Command = req.FakeOpencode
    return nil
}
```
