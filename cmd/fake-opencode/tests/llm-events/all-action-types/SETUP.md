## Preconditions
- The mock config contains one event of each `ActionType` in `llm_events`:
  think, tool_call (bash mock), message, error, done, step_start, step_finish.

## Steps
1. Run fake opencode with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_all","llm_events":[{"type":"think","text":"t1"},{"type":"tool_call","tool":"bash","mock":{"output":"hi"}},{"type":"message","text":"m1"},{"type":"error","text":"e1"},{"type":"done"},{"type":"step_start"},{"type":"step_finish"}]}`)
    return nil
}
```
