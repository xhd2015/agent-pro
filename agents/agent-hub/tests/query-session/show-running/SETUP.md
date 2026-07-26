## Preconditions
- A running session exists.

## Steps
1. Notify agent.session.started.
2. Run `agent-hub session show --runner fake-opencode --session-id s_run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_run")
    return nil
}
```
