# Scenario

**Feature**: Input is a single `ActionMessage` event with `PhaseStart`

## Preconditions
- Input is a single `ActionMessage` event with `PhaseStart`.

## Steps
1. Feed a lone `PhaseStart` event to the coalescer.
2. `PhaseStart` is never skipped.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type:  types.ActionMessage,
		ID:    "msg-1",
		Phase: types.PhaseStart,
		Text:  "hello",
	}}
	return nil
}
```
