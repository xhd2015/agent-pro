# Scenario

**Feature**: The program calls `ToOpencode` with a `message` AgentEvent

## Preconditions
- The program calls `ToOpencode` with a `message` AgentEvent.

## Steps
1. Create an AgentEvent with type `message` and text.
2. Call `ToOpencode` and print the resulting opencode events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_2",
		Type: types.ActionMessage,
		Text: "here is the result",
	}}
	req.Target = "opencode"
	req.SessionID = "sess_001"
	return nil
}
```
