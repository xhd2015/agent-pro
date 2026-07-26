## Preconditions
- The mock config contains a bash tool_call AgentEvent with a `mock` object that sets a custom exit_code and stderr.

## Steps
1. Write a mock config with mock exit_code 42.
2. Run fake-opencode and verify exit_code 42 and stderr appear in the event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_mock_exit","llm_events":[{"type":"tool_call","tool":"bash","mock":{"exit_code":42,"stderr":"custom error message"}}]}`)
    return nil
}
```
