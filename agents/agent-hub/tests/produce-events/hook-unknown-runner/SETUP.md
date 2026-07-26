## Preconditions
- hook notify is called with an unknown runner.

## Steps
1. Run `agent-hub hook notify --runner unknown --event SessionStart` with stdin payload.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"hook", "notify", "--runner", "unknown", "--event", "SessionStart"}
    req.Stdin = `{"session_id":"s1"}`
    return nil
}
```
