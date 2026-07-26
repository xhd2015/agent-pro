# Scenario

**Feature**: Sequence: `PhaseUpdate`(ID=x, text " world") → `PhaseEnd`(ID=x, text "hello world")

## Preconditions
- Sequence: `PhaseUpdate`(ID=x, text " world") → `PhaseEnd`(ID=x, text "hello world").

## Steps
1. Feed PhaseUpdate first — must not be skipped, marks ID "x" as shown.
2. Feed PhaseEnd — must be skipped.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseUpdate, Text: " world"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "hello world"},
	}
	return nil
}
```
