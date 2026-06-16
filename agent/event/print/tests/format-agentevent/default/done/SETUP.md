## Preconditions
- A `done` event (known ActionType but no special formatting).

## Steps
1. Set `req.Type = types.ActionDone`.
2. Set `req.Text = "all tasks completed"`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Type = types.ActionDone
	req.Text = "all tasks completed"
	return nil
}
```
