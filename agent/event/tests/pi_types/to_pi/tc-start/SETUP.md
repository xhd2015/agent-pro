## Preconditions
- ActionToolCall with PhaseStart produces tool_execution_start.

## Steps
1. Create ActionToolCall event with PhaseStart.
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
		Phase:     types.PhaseStart,
		Tool:      "read",
		ToolInput: map[string]any{"path": "file.txt"},
	}}
	return nil
}
```
