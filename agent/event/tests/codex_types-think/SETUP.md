# Scenario

**Feature**: The program calls `ToCodex` with a `think` AgentEvent

## Preconditions
- The program calls `ToCodex` with a `think` AgentEvent.

## Steps
1. Create an AgentEvent with type `think` and text.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_1",
		Type: types.ActionThink,
		Text: "analyzing the request",
	}}
	return nil
}
```
