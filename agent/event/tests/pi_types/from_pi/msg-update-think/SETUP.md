## Preconditions
- message_update with thinking_delta assistant event → ActionThink PhaseUpdate.

## Steps
1. Create a pi message_update event with thinking_delta.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeMessageUpdate,
		Message: &pi_types.AgentMessage{
			Role:    "assistant",
			Content: []pi_types.MessageContent{{Type: "thinking", Thinking: "think"}},
		},
		AssistantMessageEvent: &pi_types.AssistantMessageEvent{
			Type:        "thinking_delta",
			ContentIndex: 0,
			Delta:       " deeper",
		},
	}}
	return nil
}
```
