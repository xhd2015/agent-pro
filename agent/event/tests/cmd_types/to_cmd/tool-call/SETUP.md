# Scenario

**Feature**: ActionToolCall maps to an assistant tool-call block

```
# canonical tool_call event -> assistant tool-call content block
ActionToolCall(tool="bash", tool_input={"command":"echo hi"}) -> {"role":"assistant","content":[{"type":"tool-call","toolName":"bash","input":{"command":"echo hi"}}]}
```

## Preconditions
- `ToCmd` converts each `ActionToolCall` into an assistant event with a `tool-call` content block.

## Steps
1. Provide one canonical `ActionToolCall` event.
2. Verify the output contains an assistant event with a `tool-call` block.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/agent-pro/agent/event/types"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{
		{
			Type:      types.ActionToolCall,
			Tool:      "bash",
			ToolInput: map[string]any{"command": "echo hi"},
		},
	}
	return nil
}
```
