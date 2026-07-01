# Scenario

**Feature**: Sequence: `PhaseEnd`(ID=x, text "first") → `PhaseEnd`(ID=x, text "second")

## Preconditions
- Sequence: `PhaseEnd`(ID=x, text "first") → `PhaseEnd`(ID=x, text "second").

## Steps
1. Feed first PhaseEnd — not skipped (first time seeing ID "x" as PhaseEnd).
2. Feed second PhaseEnd with same ID — must be skipped (duplicate end, content already shown).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "first"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "second"},
	}
	return nil
}
```
