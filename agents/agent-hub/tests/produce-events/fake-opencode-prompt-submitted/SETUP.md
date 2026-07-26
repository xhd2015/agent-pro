## Preconditions
- A mock config with message.updated hook containing prompt text.

## Steps
1. Create mock config with message.updated hook and payload.
2. Run fake-opencode.
3. Verify agent.prompt.submitted event with prompt text.

```go
import (
    "testing"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    config := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_prompt","hooks":[{"at":"before_stdout","event":"message.updated","payload":{"message":{"text":"hello world"}}}]}`
    cfgPath := filepath.Join(req.TempDir, "mock-prompt.json")
    writeFile(t, cfgPath, config)
    req.Args = []string{"run", "--format", "json", "--mock-config", cfgPath, "hello"}
    req.Command = req.FakeOpencode
    return nil
}
```
