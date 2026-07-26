# Scenario

**Feature**: completed tool emits tool_call + tool_call_update pair

```
ActionToolCall status=completed + Output -> tool_call + tool_call_update
```

## Preconditions
- One completed ActionToolCall with Output.

## Steps
1. Provide completed ActionToolCall event.

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
		Output:     "package main",
		ToolCallID: "call_read_1",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{Status: "completed", TurnIndex: 0},
		},
	}}
	return nil
}
```
