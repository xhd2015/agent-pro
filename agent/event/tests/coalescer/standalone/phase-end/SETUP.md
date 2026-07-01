# Scenario

**Feature**: Input is a single `ActionMessage` event with `PhaseEnd`

## Preconditions
- Input is a single `ActionMessage` event with `PhaseEnd`.

## Steps
1. Feed a lone `PhaseEnd` event with text to the coalescer.
2. Since there is no prior event for this ID, it must not be skipped.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type:  types.ActionMessage,
		ID:    "msg-1",
		Phase: types.PhaseEnd,
		Text:  "hello world",
	}}
	return nil
}
```
