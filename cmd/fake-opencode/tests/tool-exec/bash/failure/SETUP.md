## Preconditions
- The mock config contains a bash tool_call AgentEvent that runs `exit 3`.

## Steps
1. Run fake-opencode with JSON output.
2. Verify the event has non-zero exit_code.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_fail","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"exit 3"}}]}`)
    return nil
}
```
