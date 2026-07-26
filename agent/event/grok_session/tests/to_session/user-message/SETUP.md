# Scenario

**Feature**: user ActionMessage converts to user_message_chunk wire

```
ActionMessage role=user -> user_message_chunk
```

## Preconditions
- One user ActionMessage event.

## Steps
1. Provide ActionMessage with role=user and text.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Role: "user",
		Text: "run ls",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
		},
	}}
	return nil
}
```
