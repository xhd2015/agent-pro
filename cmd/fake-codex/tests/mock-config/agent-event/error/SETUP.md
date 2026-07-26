## Preconditions
- The mock config contains an `error` AgentEvent with a message and nonzero exit code.

## Steps
1. Run fake Codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","exit_code":3,"stderr":"something went wrong","llm_events":[{"type":"error","text":"execution failed"}]}`)
    return nil
}
```
