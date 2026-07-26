## Preconditions
- The mock config contains an error AgentEvent and a nonzero exit code.

## Steps
1. Run fake opencode with JSON output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_error","exit_code":5,"stderr":"planned opencode failure","llm_events":[{"type":"error","text":"fake failed"}]}`)
    return nil
}
```
