## Preconditions
- An AgentEvent with Type "tool_call", tool "bash", tool_input with command, and output.

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Type:"tool_call"`.
2. Call `formatEventLine`.
3. Verify the formatted output contains the tool name.

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    event := types.AgentEvent{
        Type:  types.ActionToolCall,
        Tool:  "bash",
        Text:  "Running command",
        Output: "file1.txt\nfile2.txt",
        ToolInput: map[string]any{"command": "ls"},
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
