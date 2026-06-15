## Preconditions
- Roundtrip: ToPi then FromPi should preserve step start/finish actions.

## Steps
1. Create ActionStepStart and ActionStepFinish events.
2. Call roundtrip and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{
		{Type: types.ActionStepStart},
		{Type: types.ActionStepFinish},
	}
	return nil
}
```
