## Preconditions
- A `sleep` event (known ActionType but no special formatting).

## Steps
1. Set `req.Type = types.ActionSleep`.
2. Set `req.Text = "waiting 5000ms"`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Type = types.ActionSleep
	req.Text = "waiting 5000ms"
	return nil
}
```
