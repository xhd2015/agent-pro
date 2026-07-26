## Preconditions
- The command passes `--model openai/gpt-5`.
- A hook captures the effective model.

## Steps
1. Run fake opencode and fire a hook.

```go
import (
    "fmt"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","hook_command":%q,"hooks":[{"at":"before_stdout","event":"session.created","payload":{}}],"llm_events":[]}`, hook+" {{event}}"))
    req.Args = []string{"run", "--format", "json", "--model", "openai/gpt-5", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```

