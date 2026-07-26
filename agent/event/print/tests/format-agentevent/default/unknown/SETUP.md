## Preconditions
- An event with a completely unknown Type string.

## Steps
1. Set `req.Type = types.ActionType("custom_event")`.
2. Set `req.Text = "some arbitrary event"`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Type = types.ActionType("custom_event")
	req.Text = "some arbitrary event"
	return nil
}
```
