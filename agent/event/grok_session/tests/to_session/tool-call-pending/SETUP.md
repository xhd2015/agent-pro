# Scenario

**Feature**: pending ActionToolCall converts to tool_call wire

```
ActionToolCall status=pending -> tool_call wire
```

## Preconditions
- One pending ActionToolCall with tool_call_id.

## Steps
1. Provide ActionToolCall with status=pending.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type:       types.ActionToolCall,
		Tool:       "read",
		Text:       "README.md",
		ToolCallID: "call_read_1",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{Status: "pending", TurnIndex: 0},
		},
	}}
	return nil
}
```
