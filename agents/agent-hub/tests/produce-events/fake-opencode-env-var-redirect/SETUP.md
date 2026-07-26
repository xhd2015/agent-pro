## Preconditions
- AGENT_HUB_OPENCODE_RUNNER=fake-opencode is set.
- Mock config has runner:"opencode" with hook_command targeting opencode runner.

## Steps
1. Run fake-opencode with mock config.
2. Verify event stored with runner:"fake-opencode" (redirected).

```go
import (
    "testing"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    config := `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_redirect","hook_command":"agent-hub hook notify --runner opencode --event {{event}}","hooks":[{"at":"before_stdout","event":"session.created"}]}`
    cfgPath := filepath.Join(req.TempDir, "mock-redirect.json")
    writeFile(t, cfgPath, config)
    req.Args = []string{"run", "--format", "json", "--mock-config", cfgPath, "hello"}
    req.Command = req.FakeOpencode
    return nil
}
```
