# Scenario

**Feature**: assistant ActionMessage converts to agent_message_chunk

```
ActionMessage role=assistant -> agent_message_chunk
```

## Preconditions
- One assistant ActionMessage event.

## Steps
1. Provide ActionMessage with role=assistant.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: "Here is the answer",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
		},
	}}
	return nil
}
```
