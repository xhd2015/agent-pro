# Scenario

**Feature**: Input is a single `ActionMessage` event with `PhaseUpdate`

## Preconditions
- Input is a single `ActionMessage` event with `PhaseUpdate`.

## Steps
1. Feed a lone `PhaseUpdate` event to the coalescer.
2. `PhaseUpdate` is never skipped.

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
		Phase: types.PhaseUpdate,
		Text:  " world",
	}}
	return nil
}
```
