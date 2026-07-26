# Scenario

**Feature**: tool_execution_update → ActionToolCall PhaseUpdate

## Preconditions
- tool_execution_update → ActionToolCall PhaseUpdate.

## Steps
1. Create a pi tool_execution_update event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type:          pi_types.EventTypeToolExecUpdate,
		ToolCallID:    "call_1",
		ToolName:      "bash",
		PartialResult: "partial output",
	}}
	return nil
}
```
