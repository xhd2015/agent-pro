## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash and a command.
- No mock output, so the command executes for real.

## Steps
1. Run fake opencode with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo hello-opencode"}}]}`)
    return nil
}
```
