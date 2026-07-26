## Steps
1. Run `agent-hub integration opencode disable --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "opencode", "disable", "--help"}
    return nil
}
```
