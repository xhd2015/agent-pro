## Preconditions
- The mock config contains a bash tool_call AgentEvent with a `mock` object.

## Steps
1. Write a mock config with a bash event and `mock`.
2. Run fake-opencode and verify the mock output is used instead of real output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_mock","llm_events":[{"type":"tool_call","tool":"bash","mock":{"output":"fake mock output","exit_code":0}}]}`)
    return nil
}
```
