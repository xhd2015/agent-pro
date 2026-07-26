## Preconditions
- --from argument has bad format.

## Steps
1. Run `agent-hub replay --consumer-id c1 --from bogus`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"replay", "--consumer-id", "c1", "--from", "bogus"}
    return nil
}
```
