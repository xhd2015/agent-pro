## Preconditions
- This group tests session query operations (`session show`, `sessions`).
- Sessions are created via `agent-hub notify` events.
- `AGENT_HUB_OPENCODE_RUNNER` must be set for runner-scoped queries.

## Steps
1. Create session prerequisites via `agent-hub notify`.
2. Run `agent-hub session show` or `agent-hub sessions`.

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
