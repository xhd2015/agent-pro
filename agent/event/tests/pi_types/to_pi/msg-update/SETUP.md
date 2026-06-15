## Preconditions
- ActionMessage with PhaseUpdate produces message_update (text_delta).

## Steps
1. Create ActionMessage event with PhaseUpdate.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type:  types.ActionMessage,
		Phase: types.PhaseUpdate,
		Text:  " world",
	}}
	return nil
}
```
