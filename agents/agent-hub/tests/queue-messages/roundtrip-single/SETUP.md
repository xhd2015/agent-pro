## Preconditions
- A running session.

## Steps
1. Create session, send one exact message.
2. Pop and verify the exact message content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "srt1")
    runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "srt1", "--text", "exact message")
    return nil
}
```
