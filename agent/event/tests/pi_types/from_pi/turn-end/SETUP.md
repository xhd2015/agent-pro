# Scenario

**Feature**: turn_end → ActionStepFinish PhaseEnd

## Preconditions
- turn_end → ActionStepFinish PhaseEnd.

## Steps
1. Create a pi turn_end event with message and tool results.
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
		Type: pi_types.EventTypeTurnEnd,
		Message: &pi_types.AgentMessage{
			Role: "assistant",
			Content: []pi_types.MessageContent{{Type: "text", Text: "result"}},
		},
		ToolResults: []pi_types.ToolResultMessage{
			{ToolCallID: "tc_1", ToolName: "bash", IsError: false},
		},
	}}
	return nil
}
```
