# Scenario

**Feature**: Sequence: `PhaseStart`(ID=x, text="") → `PhaseEnd`(ID=x, text="hello")

## Preconditions
- Sequence: `PhaseStart`(ID=x, text="") → `PhaseEnd`(ID=x, text="hello").

## Steps
1. Feed PhaseStart with empty text — must not be skipped (empty start still marks ID as shown).
2. Feed PhaseEnd with text — must be skipped (ID was already shown via start).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseStart, Text: ""},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "hello"},
	}
	return nil
}
```
