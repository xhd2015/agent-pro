## Preconditions
- message_end → ActionMessage PhaseEnd.

## Steps
1. Create a pi message_end event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
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
