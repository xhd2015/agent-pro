# Scenario

**Feature**: ActionThink converts to agent_thought_chunk

```
ActionThink -> agent_thought_chunk
```

## Preconditions
- One ActionThink event.

## Steps
1. Provide ActionThink with planning text.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionThink,
		Text: "planning ls output",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
		},
	}}
	return nil
}
```
