## Preconditions
- 5 events produced.

## Steps
1. Produce 5 events.
2. Fetch with --consumer-id c1 --limit 3.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    for i := 0; i < 5; i++ {
        notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_limit_"+string(rune('0'+i)))
    }
    return nil
}
```
