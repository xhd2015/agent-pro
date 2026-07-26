# Scenario

**Feature**: The program calls `ToOpencode` with a `done` AgentEvent

## Preconditions
- The program calls `ToOpencode` with a `done` AgentEvent.

## Steps
1. Create an AgentEvent with type `done`.
2. Call `ToOpencode` and print the resulting opencode events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_done",
		Type: types.ActionDone,
	}}
	req.Target = "opencode"
	req.SessionID = "sess_001"
	return nil
}
```
