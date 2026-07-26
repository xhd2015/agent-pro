## Preconditions
- A session transitions from running to completed.

## Steps
1. Notify started.
2. Show session -> status running.
3. Notify finished.
4. Show session -> status completed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_trans")
    return nil
}
```
