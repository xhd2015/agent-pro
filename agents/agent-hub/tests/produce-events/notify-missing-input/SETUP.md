## Preconditions
- notify is called without --json or --file.

## Steps
1. Run `agent-hub notify` (no flags).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"notify"}
    return nil
}
```
