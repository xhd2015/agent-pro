## Preconditions
- No session exists.

## Steps
1. Run session show for non-existent ID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"session", "show", "--runner", "fake-opencode", "--session-id", "nosuch"}
    return nil
}
```
