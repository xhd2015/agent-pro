# Scenario

**Feature**: ActionToolCall with PhaseUpdate produces tool_execution_update

## Preconditions
- ActionToolCall with PhaseUpdate produces tool_execution_update.

## Steps
1. Create ActionToolCall event with PhaseUpdate.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type:      types.ActionToolCall,
		Phase:     types.PhaseUpdate,
		Tool:      "bash",
		ToolInput: map[string]any{"command": "ls"},
	}}
	return nil
}
```
