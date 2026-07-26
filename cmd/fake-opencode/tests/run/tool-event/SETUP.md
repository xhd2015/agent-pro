## Preconditions
- The mock config contains a tool_call AgentEvent for bash.

## Steps
1. Run fake opencode with JSON output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_tool","llm_events":[{"type":"tool_call","tool":"bash","mock":{"output":"ok"}}]}`)
    return nil
}
```
