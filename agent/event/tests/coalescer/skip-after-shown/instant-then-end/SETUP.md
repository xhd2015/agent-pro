# Scenario

**Feature**: Sequence: `PhaseInstant`(ID=x, text "instant msg") → `PhaseEnd`(ID=x, text "full msg")

## Preconditions
- Sequence: `PhaseInstant`(ID=x, text "instant msg") → `PhaseEnd`(ID=x, text "full msg").

## Steps
1. Feed PhaseInstant first — must not be skipped, marks ID "x" as shown.
2. Feed PhaseEnd — must be skipped (PhaseInstant counts as "shown").

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseInstant, Text: "instant msg"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "full msg"},
	}
	return nil
}
```
