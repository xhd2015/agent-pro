## Preconditions
- This group tests event consumption via `agent-hub fetch` and `agent-hub replay`.
- Events are produced via `agent-hub notify` or `fake-opencode run`.
- `AGENT_HUB_OPENCODE_RUNNER` may be used for runner-scoped operations.

## Steps
1. Produce events into the store.
2. Fetch events with consumer cursors.
3. Replay to reset consumer cursors.

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
