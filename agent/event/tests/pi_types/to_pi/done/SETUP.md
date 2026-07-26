# Scenario

**Feature**: ActionDone produces agent_end

## Preconditions
- ActionDone produces agent_end.

## Steps
1. Create ActionDone event.
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
		Type: types.ActionDone,
	}}
	return nil
}
```
