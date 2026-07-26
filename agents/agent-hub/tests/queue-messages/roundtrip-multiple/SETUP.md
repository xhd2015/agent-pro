## Preconditions
- A running session.

## Steps
1. Create session, send 3 messages: "A", "B", "C".
2. Pop and verify insertion order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "srt3")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "srt3", "--text", "A")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "srt3", "--text", "B")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "srt3", "--text", "C")
    return nil
}
```
