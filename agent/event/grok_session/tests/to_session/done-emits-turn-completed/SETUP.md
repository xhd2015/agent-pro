# Scenario

**Feature**: ActionDone converts to turn_completed wire

```
ActionDone -> turn_completed
```

## Preconditions
- One ActionDone event with turn_index.

## Steps
1. Provide ActionDone event.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionDone,
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{TurnIndex: 0},
		},
	}}
	return nil
}
```
