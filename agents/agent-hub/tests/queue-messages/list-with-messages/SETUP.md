## Preconditions
- A running session with 2 messages queued.

## Steps
1. Create session, send 2 messages.
2. Run `agent-hub session message list`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "slist")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "slist", "--text", "msg1")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "slist", "--text", "msg2")
    return nil
}
```
