## Preconditions
- This group tests session message queue operations (send, list, pop).
- Sessions must be created via `agent-hub notify` events before testing message commands.
- `AGENT_HUB_OPENCODE_RUNNER` must be set for runner-based operations.

## Steps
1. Create session prerequisites via `agent-hub notify`.
2. Run `agent-hub session message send/list/pop` with appropriate flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t
    return nil
}
```
