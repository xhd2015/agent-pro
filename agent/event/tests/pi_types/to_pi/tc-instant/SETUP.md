## Preconditions
- ActionToolCall with Phase="" produces tool_execution_start + tool_execution_end.

## Steps
1. Create ActionToolCall event with no phase.
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
		Tool:      "bash",
		ToolInput: map[string]any{"command": "ls"},
		Output:    "file.txt",
	}}
	return nil
}
```
