## Preconditions
- A completed session exists.

## Steps
1. Notify started + finished.
2. Run session show.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_comp")
    notifyEvent(t, req, "agent.session.finished", "fake-opencode", "s_comp")
    return nil
}
```
