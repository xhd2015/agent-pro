# Scenario

**Feature**: ToSession infers tool_call_id when missing

```
ActionToolCall without tool_call_id -> wire tool_call with generated id
```

## Preconditions
- ActionToolCall without `tool_call_id` field.

## Steps
1. Provide pending ActionToolCall with empty ToolCallID.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionToolCall,
		Tool: "read",
		Text: "README.md",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{Status: "pending", TurnIndex: 0},
		},
	}}
	return nil
}
```
