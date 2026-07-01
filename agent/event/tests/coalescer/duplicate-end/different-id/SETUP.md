# Scenario

**Feature**: Sequence: `PhaseEnd`(ID=a, text "msg a") → `PhaseEnd`(ID=b, text "msg b")

## Preconditions
- Sequence: `PhaseEnd`(ID=a, text "msg a") → `PhaseEnd`(ID=b, text "msg b").

## Steps
1. Feed first PhaseEnd with ID "a" — not skipped.
2. Feed second PhaseEnd with ID "b" — not skipped (different ID resets state, treated as standalone).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "a", Phase: types.PhaseEnd, Text: "msg a"},
		{Type: types.ActionMessage, ID: "b", Phase: types.PhaseEnd, Text: "msg b"},
	}
	return nil
}
```
