## Steps
1. Run `agent-hub integration unknown-subcommand`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "opencode", "unknown"}
    return nil
}
```
