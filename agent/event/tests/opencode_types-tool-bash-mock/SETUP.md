## Preconditions
- The program calls `ToOpencode` with a `tool_call` AgentEvent using `bash` tool and `mock` config.

## Steps
1. Create an AgentEvent with type `tool_call`, tool `bash`, and mock config.
2. Call `ToOpencode` and print the resulting opencode events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func Setup(t *testing.T, req *Request) error {
	ec := 0
	req.Events = []types.AgentEvent{{
		ID:        "evt_bash",
		Type:      types.ActionToolCall,
		Tool:      "bash",
		ToolInput: map[string]any{"command": "echo hello"},
		Mock:      &faketoolexec.MockConfig{Output: "hello", ExitCode: &ec},
	}}
	req.Target = "opencode"
	req.SessionID = "sess_001"
	return nil
}
```
