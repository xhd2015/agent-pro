# Scenario

**Feature**: ActionThink with PhaseUpdate produces message_update (thinking_delta)

## Preconditions
- ActionThink with PhaseUpdate produces message_update (thinking_delta).

## Steps
1. Create ActionThink event with PhaseUpdate.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type:  types.ActionThink,
		Phase: types.PhaseUpdate,
		Text:  " deeper thinking",
	}}
	return nil
}
```
