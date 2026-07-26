# Scenario

**Feature**: message_end → ActionMessage PhaseEnd

## Preconditions
- message_end → ActionMessage PhaseEnd.
- After fix: message_end should NOT output full text (deltas already shown via message_update).
- With no Delta available, Text must be empty.

## Steps
1. Create a pi message_end event with Content text but no assistantMessageEvent (no delta).
2. Call FromPi and marshal result.
3. Verify Text is empty, not the full Content text "Hello".

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeMessageEnd,
		Message: &pi_types.AgentMessage{
			Role:    "assistant",
			Content: []pi_types.MessageContent{{Type: "text", Text: "Hello"}},
		},
	}}
	return nil
}
```
