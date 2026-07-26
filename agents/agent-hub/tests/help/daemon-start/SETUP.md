## Steps
1. Run `agent-hub daemon start --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"daemon", "start", "--help"}
    return nil
}
```
