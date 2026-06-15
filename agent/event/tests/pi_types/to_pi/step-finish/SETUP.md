## Preconditions
- ActionStepFinish produces turn_end.

## Steps
1. Create ActionStepFinish event.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type: types.ActionStepFinish,
	}}
	return nil
}
```
