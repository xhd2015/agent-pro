## Preconditions
- --text is omitted.

## Steps
1. Run `agent-hub session message send --runner fake-opencode --session-id s1` (no --text).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    req.Args = []string{"session", "message", "send", "--runner", "fake-opencode", "--session-id", "s1"}
    return nil
}
```
