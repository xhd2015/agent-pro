# Scenario

**Feature**: appended events readable from offset zero

```
AppendEvent x3 -> ReadEvents(offset=0) -> 3 events in order
```

## Preconditions

- Session may be created implicitly by first append.
- Events are distinct messages for order verification.

## Steps

1. Set `req.Action = "append_read_start"`.
2. Provide three events with texts `alpha`, `beta`, `gamma`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Action = "append_read_start"
	req.SessionID = "sess_append"
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, Text: "alpha"},
		{Type: types.ActionMessage, Text: "beta"},
		{Type: types.ActionMessage, Text: "gamma"},
	}
	return nil
}
```