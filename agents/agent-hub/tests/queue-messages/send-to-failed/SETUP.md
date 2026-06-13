## Preconditions
- A session was created and then failed (started + failed events).

## Steps
1. Notify started + failed.
2. Send message to the failed session.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s3")
    notifyEvent(t, req, "agent.session.failed", "fake-opencode", "s3")
    return nil
}
```
