## Preconditions
- tool_execution_start → ActionToolCall PhaseStart.

## Steps
1. Create a pi tool_execution_start event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type:       pi_types.EventTypeToolExecStart,
		ToolCallID: "call_1",
		ToolName:   "bash",
		Args:       map[string]any{"command": "ls"},
	}}
	return nil
}
```
