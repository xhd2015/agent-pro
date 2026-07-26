## Preconditions
- A session was created and completed (started + finished events).

## Steps
1. Notify started + finished.
2. Send message to the completed session.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s2")
    notifyEvent(t, req, "agent.session.finished", "fake-opencode", "s2")
    return nil
}
```
