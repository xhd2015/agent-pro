# Scenario

**Feature**: agent_end → ActionDone PhaseEnd

## Preconditions
- agent_end → ActionDone PhaseEnd.

## Steps
1. Create a pi agent_end event with messages.
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
		Type: pi_types.EventTypeAgentEnd,
		Messages: []pi_types.AgentMessage{
			{Role: "assistant", Content: []pi_types.MessageContent{{Type: "text", Text: "Done"}}},
		},
	}}
	return nil
}
```
