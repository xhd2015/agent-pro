# Scenario

**Feature**: Sequence: `PhaseStart`(ID=x, text "hello") → `PhaseEnd`(ID=x, text "hello world")

## Preconditions
- Sequence: `PhaseStart`(ID=x, text "hello") → `PhaseEnd`(ID=x, text "hello world").

## Steps
1. Feed PhaseStart first — must not be skipped, marks ID "x" as shown.
2. Feed PhaseEnd — must be skipped (content already shown via start delta).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseStart, Text: "hello"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "hello world"},
	}
	return nil
}
```
