## Steps
1. Run `agent-hub daemon status --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"daemon", "status", "--help"}
    return nil
}
```
