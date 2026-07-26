## Preconditions
- The mock config uses `llm_events` with the neutral `AgentEvent` format from `agent/event/types`.
- fake-opencode converts each AgentEvent to its native opencode JSON events.
- `stdout_events` is deprecated; a stderr warning is emitted when it is present.

## Steps
1. Mark the test mode as llm-events.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=llm-events")
    return nil
}
```
