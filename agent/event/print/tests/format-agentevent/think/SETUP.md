## Preconditions
- The event Type is `think`.

## Steps
1. Set `req.Type = types.ActionThink`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Type = types.ActionThink
	return nil
}
```
