## Preconditions
- A completed session exists.

## Steps
1. Notify started + finished.
2. Run session show.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_comp")
    notifyEvent(t, req, "agent.session.finished", "fake-opencode", "s_comp")
    return nil
}
```
