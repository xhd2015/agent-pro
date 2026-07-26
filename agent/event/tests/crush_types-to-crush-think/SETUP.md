# Scenario

**Feature**: `ToCrush` converts ActionThink to a crush message with a reasoning part

## Preconditions
- `ToCrush` converts ActionThink to a crush message with a reasoning part.

## Steps
1. Create an AgentEvent with type `think` and reasoning text.
2. Call `ToCrush` and marshal the result as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_think_1",
		Type: types.ActionThink,
		Text: "thinking about the problem",
	}}
	req.Target = "crush"
	req.SessionID = "sess_crush"
	return nil
}
```
