## Preconditions
- ActionMessage with Phase="" produces message_start + message_update(text) + message_end.

## Steps
1. Create ActionMessage event with no phase.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Text: "Hello world",
	}}
	return nil
}
```
