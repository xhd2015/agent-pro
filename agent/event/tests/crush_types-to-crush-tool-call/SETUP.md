# Scenario

**Feature**: `ToCrush` converts ActionToolCall to a crush message with a tool_call part

## Preconditions
- `ToCrush` converts ActionToolCall to a crush message with a tool_call part.

## Steps
1. Create an AgentEvent with type `tool_call`, tool name, and input.
2. Call `ToCrush` and marshal the result as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_tool_1",
		Type: types.ActionToolCall,
		Tool: "bash",
		ToolInput: map[string]any{
			"command": "echo hello",
		},
	}}
	req.Target = "crush"
	req.SessionID = "sess_crush"
	return nil
}
```
