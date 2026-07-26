# Scenario

**Feature**: Sequence: `PhaseStart`(ID=x, text="hello") → `PhaseEnd`(ID=x, text="")

## Preconditions
- Sequence: `PhaseStart`(ID=x, text="hello") → `PhaseEnd`(ID=x, text="").

## Steps
1. Feed PhaseStart with text — must not be skipped.
2. Feed PhaseEnd with empty text — must be skipped (ID was shown via start, text content irrelevant).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseStart, Text: "hello"},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: ""},
	}
	return nil
}
```
