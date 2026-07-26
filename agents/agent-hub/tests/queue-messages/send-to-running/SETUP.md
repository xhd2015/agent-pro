## Preconditions
- A running session exists (created via agent.session.started).

## Steps
1. Notify agent.session.started with runner fake-opencode.
2. Send message via `agent-hub session message send --runner fake-opencode --session-id s1 --text "followup"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s1")
    return nil
}
```
