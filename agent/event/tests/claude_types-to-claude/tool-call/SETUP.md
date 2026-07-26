# Scenario

**Feature**: ActionToolCall becomes one assistant event with a tool_use block

```
# canonical tool_call -> assistant message carrying a tool_use content block
ActionToolCall{Tool:"Bash",ToolInput:{"command":"ls"}} -> {"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}
```

## Preconditions
- `ToClaude` maps `ActionToolCall` to one `assistant` event whose `message.content` has a single `tool_use` block, with `name` = `Tool` and `input` = `ToolInput`.

## Steps
1. Provide one `ActionToolCall` event with `Tool="Bash"`, `ToolInput={"command":"ls"}`.
2. Call `ToClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:        "evt_tool_1",
		Type:      types.ActionToolCall,
		Tool:      "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}}
	return nil
}
```
