## Steps
1. Run `agent-hub integration status` (no runner arg) — lists all supported runners.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "status"}
    return nil
}
```
