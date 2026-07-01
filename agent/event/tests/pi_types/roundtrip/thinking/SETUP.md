# Scenario

**Feature**: Roundtrip: ToPi then FromPi should preserve thinking content

## Preconditions
- Roundtrip: ToPi then FromPi should preserve thinking content.

## Steps
1. Create an ActionThink event with thinking text.
2. Call roundtrip and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{{
		Type: types.ActionThink,
		Text: "thinking deeply",
	}}
	return nil
}
```
