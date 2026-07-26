## Preconditions
- --consumer-id is omitted.

## Steps
1. Run `agent-hub fetch` without --consumer-id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"fetch"}
    return nil
}
```
