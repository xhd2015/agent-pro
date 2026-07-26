## Preconditions
- 1 event produced via notify.

## Steps
1. Produce 1 event.
2. Fetch with --consumer-id c1 --limit 10.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_fstart")
    return nil
}
```
