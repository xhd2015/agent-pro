## Steps
1. Run `agent-hub integration opencode uninstall --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "opencode", "uninstall", "--help"}
    return nil
}
```
