## Steps
1. Run `agent-hub partitions --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"partitions", "--help"}
    return nil
}
```
