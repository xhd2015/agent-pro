## Preconditions
- Error tool_execution_end → ActionToolCall PhaseEnd with non-zero exit code.

## Steps
1. Create a pi tool_execution_end event with isError=true.
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
		ToolCallID: "call_2",
		ToolName:   "read",
		Result:     "file not found",
		IsError:    true,
	}}
	return nil
}
```
