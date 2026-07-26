## Preconditions
- AGENT_HUB_OPENCODE_RUNNER is NOT set.
- Mock config has runner:"opencode".

## Steps
1. Run fake-opencode with mock config.
2. Verify event stored with runner:"opencode" (no redirect).

```go
import (
    "testing"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // NOTE: NOT setting AGENT_HUB_OPENCODE_RUNNER
    config := `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_default","hook_command":"agent-hub hook notify --runner opencode --event {{event}}","hooks":[{"at":"before_stdout","event":"session.created"}]}`
    cfgPath := filepath.Join(req.TempDir, "mock-default.json")
    writeFile(t, cfgPath, config)
    req.Args = []string{"run", "--format", "json", "--mock-config", cfgPath, "hello"}
    req.Command = req.FakeOpencode
    return nil
}
```
