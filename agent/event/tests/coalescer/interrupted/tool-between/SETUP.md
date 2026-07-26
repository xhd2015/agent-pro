# Scenario

**Feature**: Sequence: `PhaseEnd`(ID=a, "msg a") → `ActionToolCall`("bash") → `PhaseEnd`(ID=a, "msg a again")

## Preconditions
- Sequence: `PhaseEnd`(ID=a, "msg a") → `ActionToolCall`("bash") → `PhaseEnd`(ID=a, "msg a again").

## Steps
1. Feed first PhaseEnd — not skipped.
2. Feed ActionToolCall — not skipped, resets coalescer state.
3. Feed second PhaseEnd with same ID "a" — not skipped (state was reset by tool call).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "a", Phase: types.PhaseEnd, Text: "msg a"},
		{Type: types.ActionToolCall, ID: "", Tool: "bash", Text: "ls"},
		{Type: types.ActionMessage, ID: "a", Phase: types.PhaseEnd, Text: "msg a again"},
	}
	return nil
}
```
