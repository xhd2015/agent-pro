## Preconditions
- A running session with no messages.

## Steps
1. Create session, no messages sent.
2. Run pop.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "spop_empty")
    return nil
}
```
