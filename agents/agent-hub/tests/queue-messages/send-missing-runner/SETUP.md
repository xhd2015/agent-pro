## Preconditions
- --runner is omitted.

## Steps
1. Run `agent-hub session message send --session-id s1 --text "hi"` (no --runner).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    req.Args = []string{"session", "message", "send", "--session-id", "s1", "--text", "hi"}
    return nil
}
```
