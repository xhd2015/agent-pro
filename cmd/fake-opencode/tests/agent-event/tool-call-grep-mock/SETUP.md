## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=grep and a mock.

## Steps
1. Run fake opencode with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_mock_ae","llm_events":[{"type":"tool_call","tool":"grep","mock":{"output":"mocked grep result","exit_code":0}}]}`)
	return nil
}
```
