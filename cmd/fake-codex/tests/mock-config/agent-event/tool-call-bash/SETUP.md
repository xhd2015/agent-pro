## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash and a command.
- No mock output is set, so the command executes for real.

## Steps
1. Create a file to verify real execution.
2. Run fake Codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo hello-from-bash"}}]}`)
    return nil
}
```
