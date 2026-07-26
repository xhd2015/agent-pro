## Preconditions
- Sessions for 2 different runners exist.

## Steps
1. Create session for fake-opencode.
2. Create session for codex (via notify).
3. Run `agent-hub sessions`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_list1")
    notifyEvent(t, req, "agent.session.started", "codex", "s_list2")
    return nil
}
```
