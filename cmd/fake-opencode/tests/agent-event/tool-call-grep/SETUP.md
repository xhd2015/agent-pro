## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=grep and a pattern.
- No mock output, so grep executes for real against the repo.

## Steps
1. Create a file with known content.
2. Run fake opencode with the mock config containing a grep tool call.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeFile(t, req.MarkerPath, "UNIQUE_GREP_MARKER_AE\n")
	writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_ae","llm_events":[{"type":"tool_call","tool":"grep","tool_input":{"pattern":"UNIQUE_GREP_MARKER_AE","path":"`+req.MarkerPath+`"}}]}`)
	return nil
}
```
