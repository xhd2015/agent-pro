## Steps
1. Run `agent-hub integration` with no subcommand (no args).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration"}
    return nil
}
```
