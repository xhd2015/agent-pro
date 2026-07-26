## Preconditions
- No sessions exist.

## Steps
1. Run `agent-hub sessions` on empty store.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t.Name()
    return nil
}
```
