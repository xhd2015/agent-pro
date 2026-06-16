## Preconditions
- A step_finish event.

## Steps
1. Set `req.Type = types.ActionStepFinish`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Type = types.ActionStepFinish
	return nil
}
```
