# Scenario

**Feature**: The program calls `ToCodex` with a `tool_call` AgentEvent using `bash` tool and `mock` config

## Preconditions
- The program calls `ToCodex` with a `tool_call` AgentEvent using `bash` tool and `mock` config.

## Steps
1. Create an AgentEvent with type `tool_call`, tool `bash`, and mock config.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ec := 0
	req.Events = []types.AgentEvent{{
		ID:        "evt_bash",
		Type:      types.ActionToolCall,
		Tool:      "bash",
		ToolInput: map[string]any{"command": "echo hello"},
		Mock:      &faketoolexec.MockConfig{Output: "hello", ExitCode: &ec},
	}}
	return nil
}
```
