# Scenario

**Feature**: two ActionDone events emit two turn_completed lines

```
ActionDone(turn0) + ActionDone(turn1) -> two turn_completed wire lines
```

## Preconditions
- Two ActionDone events with turn_index 0 and 1.

## Steps
1. Provide user messages and ActionDone per turn.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{
			Type: types.ActionMessage,
			Role: "user",
			Text: "first turn",
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
			},
		},
		{
			Type: types.ActionDone,
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
			},
		},
		{
			Type: types.ActionMessage,
			Role: "user",
			Text: "second turn",
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{TurnIndex: 1},
			},
		},
		{
			Type: types.ActionDone,
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{TurnIndex: 1},
			},
		},
	}
	return nil
}
```
