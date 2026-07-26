## Preconditions
- The mock config contains a sequence of AgentEvent: think, tool_call, message.

## Steps
1. Run fake Codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"think","text":"step 1 analyze"},{"type":"tool_call","tool":"bash","tool_input":{"command":"echo step 2 execute"}},{"type":"message","text":"step 3 done"}]}`)
    return nil
}
```
