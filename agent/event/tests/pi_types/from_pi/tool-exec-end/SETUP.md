# Scenario

**Feature**: Successful tool_execution_end → ActionToolCall PhaseEnd with tool fields

## Preconditions
- Successful tool_execution_end → ActionToolCall PhaseEnd with tool fields.

## Steps
1. Create a pi tool_execution_end event with isError=false.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type:       pi_types.EventTypeToolExecEnd,
		ToolCallID: "call_1",
		ToolName:   "bash",
		Result:     "file.txt",
		IsError:    false,
	}}
	return nil
}
```
