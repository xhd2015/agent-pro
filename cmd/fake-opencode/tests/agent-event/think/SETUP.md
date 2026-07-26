## Preconditions
- The mock config contains a `think` AgentEvent with text.

## Steps
1. Run fake opencode with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_think","llm_events":[{"type":"think","text":"thinking about the problem"}]}`)
    return nil
}
```
