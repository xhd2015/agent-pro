## Preconditions
- message_start → ActionMessage PhaseStart.

## Steps
1. Create a pi message_start event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeMessageStart,
		Message: &pi_types.AgentMessage{
			Role:    "assistant",
			Content: []pi_types.MessageContent{{Type: "text", Text: "Hello"}},
		},
	}}
	return nil
}
```
