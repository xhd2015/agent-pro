# Scenario

**Feature**: failed tool emits failed tool_call_update

```
ActionToolCall status=failed -> tool_call + tool_call_update status=failed
```

## Preconditions
- One failed ActionToolCall.

## Steps
1. Provide failed ActionToolCall event.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type:       types.ActionToolCall,
		Tool:       "execute",
		Text:       "false",
		Output:     "exit code 1",
		ToolCallID: "call_exec_1",
		Extensions: &types.EventExtensions{
			GrokSession: &types.GrokSessionExtension{Status: "failed", TurnIndex: 0},
		},
	}}
	return nil
}
```
