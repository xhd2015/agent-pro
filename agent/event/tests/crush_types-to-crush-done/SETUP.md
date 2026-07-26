# Scenario

**Feature**: `ToCrush` converts ActionDone to a crush run_complete event

## Preconditions
- `ToCrush` converts ActionDone to a crush run_complete event.

## Steps
1. Create an AgentEvent with type `done`.
2. Call `ToCrush` and marshal the result as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_done_1",
		Type: types.ActionDone,
	}}
	req.Target = "crush"
	req.SessionID = "sess_crush"
	return nil
}
```
