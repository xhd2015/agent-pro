## Steps
1. Run `agent-hub integration opencode status --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "opencode", "status", "--help"}
    return nil
}
```
