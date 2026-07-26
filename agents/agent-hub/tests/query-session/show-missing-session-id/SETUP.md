## Preconditions
- --session-id is omitted.

## Steps
1. Run `agent-hub session show --runner fake-opencode` (no --session-id).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"session", "show", "--runner", "fake-opencode"}
    return nil
}
```
