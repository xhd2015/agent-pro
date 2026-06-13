## Preconditions
- The program calls `ToCodex` with a `tool_call` AgentEvent using `bash` tool and `mock` config.

## Steps
1. Create an AgentEvent with type `tool_call`, tool `bash`, and mock config.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	codex "github.com/xhd2015/agent-pro/agent/event/codex_types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func main() {
	evt := types.AgentEvent{
		ID:        "evt_bash",
		Type:      types.ActionToolCall,
		Tool:      "bash",
		ToolInput: map[string]any{"command": "echo hello"},
		Mock:      &faketoolexec.MockConfig{Output: "hello", ExitCode: intPtr(0)},
	}
	result := codex.ToCodex([]types.AgentEvent{evt})
	data, _ := json.Marshal(result)
	fmt.Println(string(data))
}

func intPtr(i int) *int { return &i }
`
	return nil
}
```
