# Scenario

**Feature**: The program calls `ToCodex` with a `message` AgentEvent

## Preconditions
- The program calls `ToCodex` with a `message` AgentEvent.

## Steps
1. Create an AgentEvent with type `message` and text.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_2",
		Type: types.ActionMessage,
		Text: "here is the result",
	}}
	return nil
}
```
