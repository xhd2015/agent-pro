# Scenario

**Feature**: Sequence: `PhaseStart`(ID=x, text "hel") → `PhaseUpdate`(ID=x, text "lo") → `PhaseEnd`(ID=x, text "hello")

## Preconditions
- Sequence: `PhaseStart`(ID=x, text "hel") → `PhaseUpdate`(ID=x, text "lo") → `PhaseEnd`(ID=x, text "hello").

## Steps
1. Feed PhaseStart — not skipped, marks ID "x" as shown.
2. Feed PhaseUpdate — not skipped.
3. Feed PhaseEnd — must be skipped (content already shown via start+update deltas).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseStart, Text: "hel"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseUpdate, Text: "lo"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: "hello"},
	}
	return nil
}
```
