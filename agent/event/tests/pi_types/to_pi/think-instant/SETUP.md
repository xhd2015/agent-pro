# Scenario

**Feature**: ActionThink with Phase="" produces message_start + message_update(thinking) + message_end

## Preconditions
- ActionThink with Phase="" produces message_start + message_update(thinking) + message_end.

## Steps
1. Create ActionThink event with no phase (instant).
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type: types.ActionThink,
		Text: "thinking about problem",
	}}
	return nil
}
```
