## Preconditions
- A running session with 2 messages.

## Steps
1. Create session, send 2 messages.
2. Run pop to drain.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "spop")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "spop", "--text", "A1")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "spop", "--text", "B2")
    return nil
}
```
