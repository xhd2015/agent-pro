# Scenario

**Feature**: `ToCrush` converts ActionError to a crush agent_event with type `error`

## Preconditions
- `ToCrush` converts ActionError to a crush agent_event with type `error`.

## Steps
1. Create an AgentEvent with type `error` and error text.
2. Call `ToCrush` and marshal the result as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "something went wrong",
	}}
	req.Target = "crush"
	req.SessionID = "sess_crush"
	return nil
}
```
